export class AssertionFailure extends Error {
  readonly actual : any
  readonly expected: any

  constructor(actual: any, expected: any) {
    super(`expected ${expected}, got ${actual}`)

    this.actual = actual
    this.expected = expected

    Object.setPrototypeOf(this, AssertionFailure.prototype)
  }
}
