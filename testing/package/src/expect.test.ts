import { expect as keelExpect } from "./expect";

describe("expect", () => {
  it("throws on failed assertion", () => {
    expect(() => keelExpect.equal(1, 2)).toThrowError();
  });

  it("does not throw on successful assertion", () => {
    expect(() => keelExpect.equal(1, 1)).not.toThrowError();
  });
});
