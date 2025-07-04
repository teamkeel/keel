import { describe, expect, test } from "vitest";
import { table, TableData } from "./table";

// Use the usage input and return the ui config response
const testTableAPI = <T extends Record<string, any>>(options: TableData<T>) => {
  return table(options as TableData<any>).uiConfig;
};

describe("table element", () => {
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
  });
});
