class InlineFile {
  constructor(key, filename, contentType, size, url) {
    this.filename = filename;
    this.contentType = contentType;
    this.size = size;
    this.key = key;
    this.public = false;
    this.url = url;
  }

  static fromDataURL(dataURL) {
    var info = dataURL.split(",")[0].split(":")[1];
    var data = dataURL.split(",")[1];

    var mime = info.split(";")[0];
    var name = info.split(";")[1].split("=")[1];
    var byteString = Buffer.from(data, "base64");
    var blob = new Blob([byteString], { type: mime });

    return new InlineFile(null, name).setData(blob);
  }

  setData(blob) {
    this._blob = blob;
    this.type = blob.type;
    this.size = blob.size;
    return this;
  }

  read() {
    if (this._blob) {
      return this._blob;
    }

    // TODO read from store

    return this._blob;
  }

  store() {
    //TODO: actually store and generate a key
    this.key = uuidv4();
  }

  toJSON() {
    return {
      filename: this.filename,
      contentType: this.contentType,
      size: this.size,
      key: this.key,
      public: false,
    };
  }
}

module.exports = {
  InlineFile,
};
