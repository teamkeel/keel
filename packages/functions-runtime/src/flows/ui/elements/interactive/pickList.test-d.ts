import { describe, expectTypeOf, test } from "vitest";
import { testFlow } from "../../../testingUtils";

describe("pick list element", () => {
  test("pick list types", () => {
    testFlow({}, async (ctx) => {
      const pick = ctx.ui.interactive.pickList("pickList", {
        supportedInputs: {
          scanner: true,
          manual: true,
        },
        duplicateHandling: "increaseQuantity",
        data: [
          {
            id: "1",
            name: "thing",
            targetQuantity: 10,
            gtins: ["1234567890123"],
          },
          {
            id: "2",
            name: "thing2",
            targetQuantity: 2,
            gtins: ["2397y49y3"],
          },
        ],
        render: (data) => ({
          id: data.id,
          targetQuantity: 1,
          title: data.name,
          barcodes: data.gtins,
        }),
      });

      const res = await ctx.ui.page("page", {
        content: [pick],
      });

      expectTypeOf(res.pickList.items).toBeArray();
      expectTypeOf(res.pickList.items).branded.toEqualTypeOf<
        {
          id: string;
          quantity: number;
          targetQuantity: number;
        }[]
      >;
    });

    test("pick list validation types", () => {
      testFlow({}, async (ctx) => {
        const pick = ctx.ui.interactive.pickList("pickList", {
          data: [
            {
              id: "1",
              name: "thing",
              targetQuantity: 10,
              gtins: ["1234567890123"],
            },
          ],
          render: (data) => ({
            id: data.id,
            targetQuantity: data.targetQuantity,
            title: data.name,
            barcodes: data.gtins,
          }),
          validate: (response) => {
            // Test that response has the correct structure
            expectTypeOf(response).toMatchTypeOf<{
              items: Array<{
                id: string;
                quantity: number;
                targetQuantity: number;
                scannedBarcodes?: string[];
              }>;
            }>();

            // Test that items array is accessible
            expectTypeOf(response.items).toBeArray();

            // Test that we can use array methods
            const totalQuantity = response.items.reduce(
              (sum, item) => sum + item.quantity,
              0
            );
            expectTypeOf(totalQuantity).toBeNumber();

            // Test that every returns boolean
            const allValid = response.items.every((item) => item.quantity > 0);
            expectTypeOf(allValid).toBeBoolean();

            // Test individual item properties
            if (response.items.length > 0) {
              expectTypeOf(response.items[0].id).toBeString();
              expectTypeOf(response.items[0].quantity).toBeNumber();
              expectTypeOf(response.items[0].targetQuantity).toBeNumber();
              expectTypeOf(response.items[0].scannedBarcodes).toEqualTypeOf<
                string[] | undefined
              >();
            }

            // Can return string for validation error
            if (totalQuantity > 20) {
              return "Total quantity cannot exceed 20 items";
            }

            // Can return boolean for success/failure
            return allValid;
          },
        });

        const res = await ctx.ui.page("page", {
          content: [pick],
        });

        // Verify response type structure
        expectTypeOf(res.pickList).toMatchTypeOf<{
          items: Array<{
            id: string;
            quantity: number;
            targetQuantity: number;
            scannedBarcodes?: string[];
          }>;
        }>();
      });
    });

    test("pick list async validation types", () => {
      testFlow({}, async (ctx) => {
        const pick = ctx.ui.interactive.pickList("pickList", {
          data: [{ id: "1", name: "thing", targetQuantity: 5, gtins: [] }],
          render: (data) => ({
            id: data.id,
            targetQuantity: data.targetQuantity,
            title: data.name,
            barcodes: data.gtins,
          }),
          validate: async (response) => {
            // Test async validate function
            expectTypeOf(response.items).toBeArray();

            const total = response.items.reduce(
              (sum, item) => sum + item.quantity,
              0
            );

            // Simulate async operation
            await new Promise((resolve) => setTimeout(resolve, 0));

            if (total > 100) {
              return "Async validation failed";
            }

            return true;
          },
        });

        await ctx.ui.page("page", {
          content: [pick],
        });
      });
    });
  });
});
