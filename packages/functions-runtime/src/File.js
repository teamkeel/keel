const {
  S3Client,
  PutObjectCommand,
  GetObjectCommand,
} = require("@aws-sdk/client-s3");
const { fromEnv } = require("@aws-sdk/credential-providers");
const { getSignedUrl } = require("@aws-sdk/s3-request-presigner");
const { useDatabase } = require("./database");
const { DatabaseError } = require("./errors");
const KSUID = require("ksuid");

const s3Client = (() => {
  if (!process.env.KEEL_FILES_BUCKET_NAME) {
    return null;
  }

  // Set in integration tests to send all AWS API calls to a test server
  // for mocking
  const endpoint = process.env.TEST_AWS_ENDPOINT;

  return new S3Client({
    region: process.env.KEEL_REGION,

    // If a test endpoint is provided then use some test credentials rather than fromEnv()
    credentials: endpoint
      ? {
          accessKeyId: "test",
          secretAccessKey: "test",
        }
      : fromEnv(),

    // If a custom endpoint is set we need to use a custom resolver. Just settng the base endpoint isn't enough for S3 as it
    // as the default resolver uses the bucket name as a sub-domain, which likely won't work with the custom endpoint.
    // By impleenting a full resolver we can force it to be the endpoint we want.
    endpointProvider: () => {
      return {
        url: new URL(endpoint),
      };
    },
  });
})();

class InlineFile {
  constructor({ filename, contentType }) {
    this._filename = filename;
    this._contentType = contentType;
    this._contents = null;
  }

  static fromDataURL(dataURL) {
    var info = dataURL.split(",")[0].split(":")[1];
    var data = dataURL.split(",")[1];
    var mime = info.split(";")[0];
    var name = info.split(";")[1].split("=")[1];
    var buffer = Buffer.from(data, "base64");
    var file = new InlineFile({ filename: name, contentType: mime });
    file.write(buffer);

    return file;
  }

  get size() {
    if (this._contents) {
      return this._contents.size;
    } else {
      return 0;
    }
  }

  get contentType() {
    return this._contentType;
  }

  get filename() {
    return this._filename;
  }

  write(buffer) {
    this._contents = new Blob([buffer]);
  }

  // Read the contents of the file. If URL is set, it will be read from the remote storage, otherwise, if dataURL is set
  // on the instance, it will return a blob with the file contents
  async read() {
    const arrayBuffer = await this._contents.arrayBuffer();
    const buffer = Buffer.from(arrayBuffer);

    return buffer;
  }

  async store(expires = null) {
    const content = await this.read();
    const key = KSUID.randomSync().string;

    await storeFile(
      content,
      key,
      this._filename,
      this._contentType,
      this.size,
      expires
    );

    return new File({
      key: key,
      size: this.size,
      filename: this.filename,
      contentType: this.contentType,
    });
  }
}

class File extends InlineFile {
  constructor(input) {
    super({ filename: input.filename, contentType: input.contentType });
    this._key = input.key;
    this._size = input.size;
  }

  static fromDbRecord({ key, filename, size, contentType }) {
    return new File({
      key: key,
      filename: filename,
      size: size,
      contentType: contentType,
    });
  }

  get size() {
    return this._size;
  }

  get key() {
    return this._key;
  }

  async read() {
    if (this._contents) {
      const arrayBuffer = await this._contents.arrayBuffer();
      return Buffer.from(arrayBuffer);
    }

    if (s3Client) {
      const params = {
        Bucket: process.env.KEEL_FILES_BUCKET_NAME,
        Key: "files/" + this.key,
      };
      const command = new GetObjectCommand(params);
      const response = await s3Client.send(command);
      const blob = await response.Body.transformToByteArray();
      return Buffer.from(blob);
    }

    // default to db storage
    const db = useDatabase();

    try {
      let query = db
        .selectFrom("keel_storage")
        .select("data")
        .where("id", "=", this.key);

      const row = await query.executeTakeFirstOrThrow();
      return row.data;
    } catch (e) {
      throw new DatabaseError(e);
    }
  }

  async store(expires = null) {
    // Only necessary to store the file if the contents have been changed
    if (this._contents) {
      const contents = await this.read();
      await storeFile(
        contents,
        this.key,
        this.filename,
        this.contentType,
        expires
      );
    }
    return this;
  }

  async getPresignedUrl() {
    if (s3Client) {
      const command = new GetObjectCommand({
        Bucket: process.env.KEEL_FILES_BUCKET_NAME,
        Key: "files/" + this.key,
        ResponseContentDisposition: "inline",
      });

      const url = await getSignedUrl(s3Client, command, { expiresIn: 60 * 60 });

      return new URL(url);
    } else {
      const contents = await this.read();
      const dataurl = `data:${this.contentType};name=${
        this.filename
      };base64,${contents.toString("base64")}`;
      return new URL(dataurl);
    }
  }

  toDbRecord() {
    return {
      key: this.key,
      filename: this.filename,
      contentType: this.contentType,
      size: this.size,
    };
  }

  toJSON() {
    return {
      key: this.key,
      filename: this.filename,
      contentType: this.contentType,
      size: this.size,
    };
  }
}

async function storeFile(contents, key, filename, contentType, expires) {
  if (s3Client) {
    const params = {
      Bucket: process.env.KEEL_FILES_BUCKET_NAME,
      Key: "files/" + key,
      Body: contents,
      ContentType: contentType,
      ContentDisposition: `attachment; filename="${encodeURIComponent(
        filename
      )}"`,
      Metadata: {
        filename: filename,
      },
      ACL: "private",
    };

    if (expires) {
      if (expires instanceof Date) {
        params.Expires = expires;
      } else {
        console.warn("Invalid expires value. Skipping Expires parameter.");
      }
    }

    const command = new PutObjectCommand(params);
    try {
      await s3Client.send(command);
    } catch (error) {
      console.error("Error uploading file:", error);
      throw error;
    }
  } else {
    const db = useDatabase();

    try {
      let query = db
        .insertInto("keel_storage")
        .values({
          id: key,
          filename: filename,
          content_type: contentType,
          data: contents,
        })
        .onConflict((oc) =>
          oc
            .column("id")
            .doUpdateSet(() => ({
              filename: filename,
              content_type: contentType,
              data: contents,
            }))
            .where("keel_storage.id", "=", key)
        )
        .returningAll();

      await query.execute();
    } catch (e) {
      throw new DatabaseError(e);
    }
  }
}

module.exports = {
  InlineFile,
  File,
};
