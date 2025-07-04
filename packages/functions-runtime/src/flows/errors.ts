export class NonRetriableError extends Error {
  constructor(message?: string) {
    super(message);
  }
}
