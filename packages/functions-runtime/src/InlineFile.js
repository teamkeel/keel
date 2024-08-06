const { S3Client, PutObjectCommand } = require("@aws-sdk/client-s3");
const { fromEnv } = require("@aws-sdk/credential-providers");
const { useDatabase } = require("./database");
const { DatabaseError } = require("./errors");
const KSUID = require("ksuid");

class InlineFile {
  constructor(filename, contentType, size, url) {
    this.filename = filename;
    this.contentType = contentType;
    this.size = size;
    this.url = url;
  }

  // Create an InlineFile instance from a given json object.
  static fromObject(obj) {
    if (obj.dataURL) {
      var file = InlineFile.fromDataURL(obj.dataURL);
      file._dataURL = obj.dataURL;
      return file;
    }

    return new InlineFile(obj.filename, obj.contentType, obj.size, obj.url);
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
  read() {
    if (this.url) {
      // TODO: read from store
    }

    if (this._dataURL) {
      var data = this._dataURL.split(",")[1];
      return Buffer.from(data, "base64");
    }
  }

  async store(expires = null) {
    const content = this.read();
    this.key = KSUID.randomSync().string;

    if (isS3Storage()) {
      const s3Client = new S3Client({
        credentials: fromEnv(),
        region: process.env.KEEL_REGION,
      });

      const params = {
        Bucket: process.env.KEEL_FILES_BUCKET_NAME,
        Key: this.key,
        Body: content,
        ContentType: this.contentType,
        Metadata: {
          filename: this.filename,
        },
      };

      if (expires)  {
        params.Expires = expires
      }

      const command = new PutObjectCommand(params);
      try {
        const result = await s3Client.send(command);
        console.log(`File uploaded successfully. ETag: ${result.ETag}`);
        return {
          key: this.key,
          size: this.size,
          filename: this.filename,
          contentType: this.contentType,
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
    };
  }
}

module.exports = {
  InlineFile,
};

function isS3Storage() {
  return "KEEL_FILES_BUCKET_NAME" in process.env;
}
