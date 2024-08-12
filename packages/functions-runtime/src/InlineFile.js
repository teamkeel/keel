const {
  S3Client,
  PutObjectCommand,
  GetObjectCommand,
} = require("@aws-sdk/client-s3");
const { fromEnv } = require("@aws-sdk/credential-providers");
const { useDatabase } = require("./database");
const { DatabaseError } = require("./errors");
const KSUID = require("ksuid");

class InlineFile {
  constructor(filename, contentType, size, url, key, pub) {
    this.filename = filename;
    this.contentType = contentType;
    this.size = size;
    this.url = url;
    this.key = key;
    this.public = pub || false;
  }

  // Create an InlineFile instance from a given json object.
  static fromObject(obj) {
    if (obj.dataURL) {
      var file = InlineFile.fromDataURL(obj.dataURL);
      file._dataURL = obj.dataURL;
      return file;
    }

    return new InlineFile(
      obj.filename,
      obj.contentType,
      obj.size,
      obj.url,
      obj.key,
      obj.public
    );
  }

  // Create an InlineFile instance from a given dataURL
  static fromDataURL(dataURL) {
    var info = dataURL.split(",")[0].split(":")[1];
    var data = dataURL.split(",")[1];

    var mime = info.split(";")[0];
    var name = info.split(";")[1].split("=")[1];
    var byteString = Buffer.from(data, "base64");
    var blob = new Blob([byteString], { type: mime });

    var file = new InlineFile(name, mime, blob.size);
    file._dataURL = dataURL;
    return file;
  }

  // Read the contents of the file. If URL is set, it will be read from the remote storage, otherwise, if dataURL is set
  // on the instance, it will return a blob with the file contents
  async read() {
    if (this._dataURL) {
      var data = this._dataURL.split(",")[1];
      return Buffer.from(data, "base64");
    }

    // if we don't have a key nor a dataURL, this inline file has no data
    if (!this.key) {
      throw new Error("invalid file data");
    }

    if (isS3Storage()) {
      const s3Client = new S3Client({
        credentials: fromEnv(),
        region: process.env.KEEL_REGION,
      });

      const params = {
        Bucket: process.env.KEEL_FILES_BUCKET_NAME,
        Key: "files/" + this.key,
      };
      const command = new GetObjectCommand(params);
      const response = await s3Client.send(command);
      const blob = response.Body.transformToByteArray();
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
    const content = await this.read();
    this.key = KSUID.randomSync().string;

    if (isS3Storage()) {
      const s3Client = new S3Client({
        credentials: fromEnv(),
        region: process.env.KEEL_REGION,
      });

      const params = {
        Bucket: process.env.KEEL_FILES_BUCKET_NAME,
        Key: "files/" + this.key,
        Body: content,
        ContentType: this.contentType,
        Metadata: {
          filename: this.filename,
        },
        ACL: this.public ? "public-read" : "private",
      };

      if (expires) {
        params.Expires = expires;
      }

      const command = new PutObjectCommand(params);
      try {
        await s3Client.send(command);

        return {
          key: this.key,
          size: this.size,
          filename: this.filename,
          contentType: this.contentType,
          public: this.public,
        };
      } catch (error) {
        console.error("Error uploading file:", error);
        throw error;
      }
    }

    // default to db storage
    const db = useDatabase();

    try {
      let query = db.insertInto("keel_storage").values({
        id: this.key,
        filename: this.filename,
        content_type: this.contentType,
        data: content,
      });

      await query.returningAll().executeTakeFirstOrThrow();
      return {
        key: this.key,
        size: this.size,
        filename: this.filename,
        contentType: this.contentType,
      };
    } catch (e) {
      throw new DatabaseError(e);
    }
  }

  toJSON() {
    return {
      __typename: "InlineFile",
      dataURL: this._dataURL,
      filename: this.filename,
      contentType: this.contentType,
      size: this.size,
      url: this.url,
      public: this.public,
    };
  }
}

module.exports = {
  InlineFile,
};

function isS3Storage() {
  return "KEEL_FILES_BUCKET_NAME" in process.env;
}
