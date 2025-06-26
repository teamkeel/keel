import { describe, expect, test } from "vitest";
import { dataGridInput, DataGridOptions } from "./dataGrid";

// Use the usage input and return the ui config response
const testDataGridAPI = (options: Parameters<typeof dataGridInput>[1]) => {
  return dataGridInput("name", options).uiConfig;
};

describe("data grid input element", () => {
  describe("ui config", () => {
    test("single column", () => {
      const res = testDataGridAPI({
        data: [
          {
            name: "John",
            age: 20,
            email: "john@example.com",
          },
        ],
        columns: [
          {
            key: "name",
            editable: true,
            type: "text",
          },
        ],
      });

      expect(res.data).toEqual([
        {
          name: "John",
        },
      ]);

      expect(res.columns).toEqual([
        {
          key: "name",
          label: "name",
          index: 0,
          type: "text",
          editable: true,
        },
      ]);
    });
    test("inferred columns", () => {
      const res = testDataGridAPI({
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
          key: "name",
          label: "Name",
          index: 0,
          type: "text",
          editable: true,
        },
        {
          key: "age",
          label: "Age",
          index: 1,
          type: "number",
          editable: true,
        },
        {
          key: "email",
          label: "Email",
          index: 2,
          type: "text",
          editable: true,
        },
      ]);
    });
  });
});
