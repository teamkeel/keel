import { expect as keelExpect } from "./expect";

describe("expect", () => {
  describe("toEqual", () => {
    it("throws on failed assertion", () => {
      expect(() => keelExpect(1).toEqual(2)).toThrowError();
    });

    it("does not throw on successful assertion", () => {
      expect(() => keelExpect(1).toEqual(1)).not.toThrowError();
    });
  });

  describe("notToEqual", () => {
    it("throws on failed assertion", () => {
      expect(() => keelExpect(2).notToEqual(2)).toThrowError();
    });

    it("does not throw on successful assertion", () => {
      expect(() => keelExpect(1).notToEqual(2)).not.toThrowError();
    });
  });

  describe("toHaveError", () => {
    it("throws on no errors", () => {
      const actionResult = {};

      expect(() => keelExpect(actionResult).toHaveError()).toThrowError();
    });

    it("throws when there are errors but there isnt a match", () => {
      const actionResult = {
        errors: [
          {
            message: "something went wrong",
          },
        ],
      };

      const error = {
        message: "unexpected error",
      };

      expect(() => keelExpect(actionResult).toHaveError(error)).toThrowError();
    });

    it("does not throw when there is a match", () => {
      const actionResult = {
        errors: [
          {
            message: "something went wrong",
          },
        ],
      };

      const error = {
        message: "something went wrong",
      };

      expect(() =>
        keelExpect(actionResult).toHaveError(error)
      ).not.toThrowError();
    });
  });

  describe("notToHaveError", () => {
    it("throws on matching error", () => {
      const actionResult = {
        errors: [
          {
            message: "foo",
          },
        ],
      };

      expect(() =>
        keelExpect(actionResult).notToHaveError({ message: "foo" })
      ).toThrowError();
    });

    it("does not throw when there is no match", () => {
      const actionResult = {
        errors: [
          {
            message: "something went wrong",
          },
        ],
      };

      const error = {
        message: "another error",
      };

      expect(() =>
        keelExpect(actionResult).notToHaveError(error)
      ).not.toThrowError();
    });
  });

  describe("toHaveAuthorizationError", () => {
    it("does not throw when there is an auth error", () => {
      const actionResult = {
        errors: [
          {
            message: "not authorized to access this operation",
          },
        ],
      };

      expect(() =>
        keelExpect(actionResult).toHaveAuthorizationError()
      ).not.toThrowError();
    });

    it("does not throw when there is a different error", () => {
      const actionResult = {
        errors: [
          {
            message: "oops something went wrong",
          },
        ],
      };

      expect(() =>
        keelExpect(actionResult).toHaveAuthorizationError()
      ).toThrowError();
    });
  });

  describe("toBeEmpty", () => {
    it("throws when not null / undefined", () => {
      const v = {
        foo: "bar",
      };

      expect(() => keelExpect(v).toBeEmpty()).toThrowError();
    });

    it("does not throw when null", () => {
      expect(() => keelExpect(null).toBeEmpty()).not.toThrowError();
    });

    it("does not throw when undefined", () => {
      expect(() => keelExpect(undefined).toBeEmpty()).not.toThrowError();
    });
  });

  describe("notToBeEmpty", () => {
    it("throws when null / undefined", () => {
      expect(() => keelExpect(null).notToBeEmpty()).toThrowError();
      expect(() => keelExpect(undefined).notToBeEmpty()).toThrowError();
    });

    it("does not throw when undefined", () => {
      const v = {
        foo: "bar",
      };
      expect(() => keelExpect(v).notToBeEmpty()).not.toThrowError();
    });
  });

  describe("toContain", () => {
    it("throws when there is no match", () => {
      const v = [
        {
          foo: "bar",
        },
      ];

      const lookup = {
        foo: "foo",
      };

      expect(() => keelExpect(v).toContain(lookup)).toThrowError();
    });

    it("does not throw when there is a match", () => {
      const v = [
        {
          foo: "bar",
        },
      ];

      const lookup = {
        foo: "bar",
      };

      expect(() => keelExpect(v).toContain(lookup)).not.toThrowError();
    });
  });

  describe("notToContain", () => {
    it("does not throw when there is no match", () => {
      const v = [
        {
          foo: "bar",
        },
      ];

      const lookup = {
        foo: "foo",
      };

      expect(() => keelExpect(v).notToContain(lookup)).not.toThrowError();
    });

    it("throws when there is a match", () => {
      const v = [
        {
          foo: "bar",
        },
      ];

      const lookup = {
        foo: "bar",
      };

      expect(() => keelExpect(v).notToContain(lookup)).toThrowError();
    });
  });
});
