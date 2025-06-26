import { describe, expectTypeOf, test } from "vitest";
import { testFlow } from "../../../testingUtils";

describe("data grid input element", () => {
  test("all columns", () => {
    testFlow({}, async (ctx) => {
      const thing: string = "foo";
      const res = await ctx.ui.page("page", {
        content: [
          ctx.ui.inputs.dataGrid("name", {
            data: [
              {
                foo: thing,
                bar: thing,
                baz: true,
                moo: 123,
              },
              {
                foo: thing,
                bar: "another thing",
                baz: false,
                moo: 432,
              },
            ],
          }),
        ],
      });

      expectTypeOf(res.name).toBeArray();
      expectTypeOf(res.name[0]).branded.toEqualTypeOf<{
        foo: string;
        bar: string;
        baz: boolean;
        moo: number;
      }>;
    });
  });
  test("define columns with inferred types", () => {
    testFlow({}, async (ctx) => {
      const thing: string = "foo";
      const res = await ctx.ui.page("page", {
        content: [
          ctx.ui.inputs.dataGrid("name", {
            data: [
              {
                foo: thing,
                bar: thing,
                baz: 1,
                moo: 123,
              },
              {
                foo: thing,
                bar: thing,
                baz: 1,
                moo: 123,
              },
            ],
            columns: [
              {
                key: "foo",
                editable: true,
              },
            ],
          }),
        ],
      });

      expectTypeOf(res.name).toBeArray();
      expectTypeOf(res.name).branded.toEqualTypeOf<
        {
          foo: string;
        }[]
      >();
    });
  });
  test("define one columns with mapped types", () => {
    testFlow({}, async (ctx) => {
      const thing: string = "foo";
      const res = await ctx.ui.page("page", {
        content: [
          ctx.ui.inputs.dataGrid("name", {
            data: [
              {
                foo: thing,
                bar: thing,
                baz: false,
                moo: 123,
              },
              {
                foo: thing,
                bar: thing,
                baz: false,
                moo: 123,
              },
            ],
            columns: [
              {
                key: "foo",
                editable: true,
                type: "number",
              },
            ],
          }),
        ],
      });

      expectTypeOf(res.name).toBeArray();
      expectTypeOf(res.name[0]).branded.toEqualTypeOf<{
        foo: number;
      }>();
    });
    test("define multiple columns with mapped types", () => {
      testFlow({}, async (ctx) => {
        const thing: string = "foo";
        const res = await ctx.ui.page("page", {
          content: [
            ctx.ui.inputs.dataGrid("name", {
              data: [
                {
                  foo: thing,
                  bar: thing,
                  baz: false,
                  moo: 123,
                },
                {
                  foo: thing,
                  bar: thing,
                  baz: false,
                  moo: 123,
                },
              ],
              columns: [
                {
                  key: "foo",
                  editable: true,
                  type: "number",
                },
                {
                  key: "bar",
                  editable: true,
                  type: "boolean",
                },
                {
                  key: "baz",
                  editable: true,
                  type: "id",
                },
                {
                  key: "moo",
                  editable: true,
                  type: "text",
                },
              ],
            }),
          ],
        });

        expectTypeOf(res.name).toBeArray();
        expectTypeOf(res.name[0]).branded.toEqualTypeOf<{
          foo: number;
          bar: boolean;
          baz: string;
          moo: string;
        }>();
      });
    });
  });
});
