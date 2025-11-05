import {
  S3Client,
  PutObjectCommand,
  GetObjectCommand,
  PutObjectCommandInput,
} from "@aws-sdk/client-s3";
import { fromEnv } from "@aws-sdk/credential-providers";
import { getSignedUrl } from "@aws-sdk/s3-request-presigner";
import { useDatabase } from "./database";
import { DatabaseError } from "./errors";
import KSUID from "ksuid";

type MimeType =
  | "application/json"
  | "application/gzip"
  | "application/pdf"
  | "application/rtf"
  | "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
  | "application/vnd.openxmlformats-officedocument.presentationml.presentation"
  | "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
  | "application/vnd.ms-excel"
  | "application/vnd.ms-powerpoint"
  | "application/msword"
  | "application/zip"
  | "application/xml"
  | "application/x-7z-compressed"
  | "application/x-tar"
  | "image/gif"
  | "image/jpeg"
  | "image/svg+xml"
  | "image/png"
  | "text/html"
  | "text/csv"
  | "text/javascript"
  | "text/plain"
  | "text/calendar"
  | (string & {});

// Type Declarations
export type InlineFileConstructor = {
  filename: string;
  contentType: MimeType;
};

export type FileDbRecord = {
  key: string;
  filename: string;
  contentType: string;
  size: number;
};

// Implementation
const s3Client: S3Client | null = (() => {
  if (!process.env.KEEL_FILES_BUCKET_NAME) {
    return null;
  }

  const endpoint = process.env.KEEL_S3_ENDPOINT;
  if (endpoint) {
    return new S3Client({
      region: process.env.KEEL_REGION,
      credentials: {
        accessKeyId: "keelstorage",
        secretAccessKey: "keelstorage",
      },
      endpoint: endpoint,
    });
  }

  const testEndpoint = process.env.TEST_AWS_ENDPOINT;
  if (testEndpoint) {
    return new S3Client({
      region: process.env.KEEL_REGION,
      credentials: {
        accessKeyId: "test",
        secretAccessKey: "test",
      },
      endpointProvider: () => {
        return {
          url: new URL(testEndpoint),
        };
      },
    });
  }

  return new S3Client({
    region: process.env.KEEL_REGION,
    credentials: fromEnv(),
  });
})();

export class InlineFile {
  protected _filename: string;
  protected _contentType: MimeType;
  protected _contents: Blob | null;

  constructor(input: InlineFileConstructor) {
    this._filename = input.filename;
    this._contentType = input.contentType;
    this._contents = null;
  }

  static fromDataURL(dataURL: string): InlineFile {
    const info = dataURL.split(",")[0].split(":")[1];
    const data = dataURL.split(",")[1];
    const mimeType = info.split(";")[0];
    const name = info.split(";")[1].split("=")[1] || "file";
    const buffer = Buffer.from(data, "base64");
    const file = new InlineFile({ filename: name, contentType: mimeType });
    file.write(buffer);

    return file;
  }

  // Gets size of the file's contents in bytes
  get size(): number {
    if (this._contents) {
      return this._contents.size;
    }
    return 0;
  }

  // Gets the media type of the file contents
  get contentType(): string {
    return this._contentType;
  }

  // Gets the name of the file
  get filename(): string {
    return this._filename;
  }

  // Write the files contents from a buffer
  write(buffer: Buffer): void {
    this._contents = new Blob([buffer]);
  }

  // Reads the contents of the file as a buffer
  async read(): Promise<Buffer> {
    if (!this._contents) {
      throw new Error("No contents to read");
    }
    const arrayBuffer = await this._contents.arrayBuffer();
    return Buffer.from(arrayBuffer);
  }

  // Persists the file
  async store(expires: Date | null = null): Promise<File> {
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

export class File extends InlineFile {
  private _key: string;
  private _size: number;

  constructor(input: Partial<FileDbRecord>) {
    super({
      filename: input.filename || "",
      contentType: input.contentType || "",
    });
    this._key = input.key || "";
    this._size = input.size || 0;
  }

  // Creates a new instance from the database record
  static fromDbRecord(input: FileDbRecord): File {
    return new File({
      key: input.key,
      filename: input.filename,
      size: input.size,
      contentType: input.contentType,
    });
  }

  get size(): number {
    return this._size;
  }

  // Gets the stored key
  get key(): string {
    return this._key;
  }

  get isPublic(): boolean {
    return false; // Implement as needed
  }

  async read(): Promise<Buffer> {
    if (this._contents) {
      const arrayBuffer = await this._contents.arrayBuffer();
      return Buffer.from(arrayBuffer);
    }

    if (!s3Client) {
      throw new Error("S3 client is required");
    }

    const params = {
      Bucket: process.env.KEEL_FILES_BUCKET_NAME,
      Key: "files/" + this.key,
    };
    const command = new GetObjectCommand(params);
    const response = await s3Client.send(command);
    const blob = await response.Body!.transformToByteArray();
    return Buffer.from(blob);
  }

  async store(expires: Date | null = null): Promise<File> {
    if (this._contents) {
      const contents = await this.read();
      await storeFile(
        contents,
        this.key,
        this.filename,
        this.contentType,
        this.size,
        expires
      );
    }
    return this;
  }

  // Generates a presigned download URL
  async getPresignedUrl(): Promise<URL> {
    if (!s3Client) {
      throw new Error("S3 client is required");
    }

    const command = new GetObjectCommand({
      Bucket: process.env.KEEL_FILES_BUCKET_NAME,
      Key: "files/" + this.key,
      ResponseContentDisposition: "inline",
    });

    const url = await getSignedUrl(s3Client, command, { expiresIn: 60 * 60 });

    return new URL(url);
  }

  // Persists the file
  toDbRecord(): FileDbRecord {
    return {
      key: this.key,
      filename: this.filename,
      contentType: this.contentType,
      size: this.size,
    };
  }

  toJSON(): FileDbRecord {
    return this.toDbRecord();
  }
}

async function storeFile(
  contents: Buffer,
  key: string,
  filename: string,
  contentType: string,
  size: number,
  expires: Date | null
): Promise<void> {
  if (!s3Client) {
      throw new Error("S3 client is required");
    }

  const params: PutObjectCommandInput = {
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
}

export type FileWriteTypes = InlineFile | File;
