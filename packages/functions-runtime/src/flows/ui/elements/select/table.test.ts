import { describe, expect, test } from "vitest";
import { SelectMode, selectTable, TableOptions } from "./table";

// Use the usage input and return the ui config response
const testTableAPI = <T extends Record<string, any>, M extends SelectMode>(
  options: TableOptions<T, M>
) => {
  return selectTable("name", options as TableOptions<any>).uiConfig;
};

describe("select.table element", () => {
  describe("ui config", () => {
    test("all columns", () => {
      const res = testTableAPI({
        data: [
          {
            name: "John",
            age: 20,
            email: "john@example.com",
          },
        ],
      });

      expect(res.data).toEqual([
        {
          name: "John",
          age: 20,
          email: "john@example.com",
        },
      ]);

      expect(res.columns).toEqual([
        {
          name: "name",
          index: 0,
        },
        {
          name: "age",
          index: 1,
        },
        {
          name: "email",
          index: 2,
        },
      ]);
    });
    test("columns can be provided", () => {
      const res = testTableAPI({
        data: [
          {
            name: "John",
            age: 20,
          },
        ],
        columns: ["name", "age"],
      });

      expect(res.data).toEqual([
        {
          name: "John",
          age: 20,
        },
      ]);

      expect(res.columns).toEqual([
        {
          name: "name",
          index: 0,
        },
        {
          name: "age",
          index: 1,
        },
      ]);
    });

    test("single select mode", () => {
      const res = testTableAPI({
        data: [
          {
            name: "John",
            age: 20,
            email: "john@example.com",
          },
        ],
        mode: "single",
      });
      expect(res.mode).toEqual("single");
    });
    test("multi select mode", () => {
      const res = testTableAPI({
        data: [
          {
            name: "John",
            age: 20,
            email: "john@example.com",
          },
        ],
        mode: "multi",
      });
      expect(res.mode).toEqual("multi");
    });
  });
});
