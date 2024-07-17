class InlineFile {
  constructor(key, filename, contentType, size, url) {
    this.filename = filename;
    this.contentType = contentType;
    this.size = size;
    this.key = key;
    this.public = false;
    this.url = url;
  }

  // Create an InlineFile instance from a given json object.
  static fromObject(obj) {
    if (obj.dataURL) {
      var file = InlineFile.fromDataURL(obj.dataURL);
      file._dataURL = obj.dataURL;
      return file;
    }

    return new InlineFile(
      obj.key,
      obj.filename,
      obj.contentType,
      obj.size,
      obj.url
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

    var file = new InlineFile(null, name, mime, blob.size);
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
      var byteString = Buffer.from(data, "base64");
      return new Blob([byteString], { type: this.contentType });
    }
  }

  store() {
    //TODO: actually store and generate a key
    this.key = uuidv4();
  }

  toJSON() {
    return {
      __typename: "InlineFile",
      dataURL: this._dataURL,
      filename: this.filename,
      contentType: this.contentType,
      size: this.size,
      key: this.key,
      url: this.url,
      public: false,
    };
  }
}

module.exports = {
  InlineFile,
};
