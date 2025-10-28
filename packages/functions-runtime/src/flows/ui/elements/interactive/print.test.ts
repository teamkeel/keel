import { describe, expect, test } from "vitest";
import { print, UiElementPrint } from "./print";
import { NullableHardware } from "../../..";

// Use the usage input and return the ui config response
type PrintOptions<T extends NullableHardware> = Parameters<
  UiElementPrint<T>
>[0];

const testPrintAPI = async <T extends NullableHardware = undefined>(
  options: PrintOptions<T>
) => {
  return (await print(options as PrintOptions<T>)).uiConfig;
};

describe("print element", () => {
  describe("ui config", () => {
    test("printer routing", async () => {
      const res = await testPrintAPI<{ printers: [{ name: "myPrinter" }] }>({
        jobs: [
          {
            type: "zpl",
            data: "test",
            printer: "myPrinter",
          },
        ],
      });

      expect(res.data[0].printer).toEqual("myPrinter");
    });
  });
});
