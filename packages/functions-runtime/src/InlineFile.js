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
  constructor(input) {
   this._filename = input.filename;
    this._contentType = input.contentType;
    this._contents = null;
  }

  // // Create an InlineFile instance from a given json object.
  // static fromObject(obj) {
  //   if (obj.dataURL) {
  //     var file = InlineFile.fromDataURL(obj.dataURL);
  //     file._dataURL = obj.dataURL;
  //     return file;
  //   }

  //   return new InlineFile(
  //     obj.filename,
  //     obj.contentType,
  //     obj.size,
  //     obj.url,
  //     obj.key,
  //     obj.public
  //   );
  // }


  // Create an InlineFile instance from a given dataURL
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
      return 1;
    } else {
      return 0;
    }
  }

  // contentType() {
  //   return this._contentType;
  // }

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

    // if (this._dataURL) {
    //   var data = this._dataURL.split(",")[1];
    //   return Buffer.from(data, "base64");
    // }

    // // if we don't have a key nor a dataURL, this inline file has no data
    // if (!this.key) {
    //   throw new Error("invalid file data");
    // }

    // if (isS3Storage()) {
    //   const s3Client = new S3Client({
    //     credentials: fromEnv(),
    //     region: process.env.KEEL_REGION,
    //   });

    //   const params = {
    //     Bucket: process.env.KEEL_FILES_BUCKET_NAME,
    //     Key: "files/" + this.key,
    //   };
    //   const command = new GetObjectCommand(params);
    //   const response = await s3Client.send(command);
    //   const blob = response.Body.transformToByteArray();
    //   return Buffer.from(blob);
    // }

    // // default to db storage
    // const db = useDatabase();

    // try {
    //   let query = db
    //     .selectFrom("keel_storage")
    //     .select("data")
    //     .where("id", "=", this.key);

    //   const row = await query.executeTakeFirstOrThrow();
    //   return row.data;
    // } catch (e) {
    //   throw new DatabaseError(e);
    // }
  }

  async store(expires = null, isPublic = false) {
    const content = await this.read();
    const key = KSUID.randomSync().string;

    if (isS3Storage()) {
      const s3Client = new S3Client({
        credentials: fromEnv(),
        region: process.env.KEEL_REGION,
      });

      const params = {
        Bucket: process.env.KEEL_FILES_BUCKET_NAME,
        Key: "files/" + key,
        Body: content,
        ContentType: this._contentType,
        Metadata: {
          filename: this._filename,
        },
        ACL: isPublic ? "public-read" : "private",
      };

      if (expires) {
        params.Expires = expires;
      }

      const command = new PutObjectCommand(params);
      try {
        await s3Client.send(command);

        return new StoredFile({
          key: key,
          size: this.size,
          filename: this._filename,
          contentType: this._contentType,
          isPublic: isPublic,
        })

        // return {
        //   key: this.key,
        //   size: this.size,
        //   filename: this.filename,
        //   contentType: this.contentType,
        //   public: public,
        // };
      } catch (error) {
        console.error("Error uploading file:", error);
        throw error;
      }
    }

    // default to db storage
    const db = useDatabase();

    try {
      let query = db.insertInto("keel_storage").values({
        id: key,
        filename: this._filename,
        content_type: this._contentType,
        data: content,
      });

      await query.returningAll().executeTakeFirstOrThrow();
      // return {
      //   //key: this.key,
      //   size: this.size,
      //   filename: this._filename,
      //   contentType: this._contentType,
      // };

      return new StoredFile({
        key: key,
        filename: this._filename,
        contentType: this._contentType,
        isPublic: isPublic,
      })

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


class StoredFile extends InlineFile {
  _key;
  _isPublic;

  constructor(input) {
    super({ filename: input.filename, contentType: input.contentType });
     this._key  = input.key
     this._isPublic = input.isPublic;
   }

   get key() {
    return this._key;
   }

   get isPublic() {
    return this._isPublic;
   }

  toColumn() {
    return {
      key: this._key,
      filename: this._filename,
      contentType: this._contentType,
      size: this.size
    }
  }
}

module.exports = {
  InlineFile,
  StoredFile,
};

function isS3Storage() {
  return "KEEL_FILES_BUCKET_NAME" in process.env;
}
