import { expectTypeOf, test } from "vitest";
import { testFlowContext } from "../../../testingUtils";

test("grid elements types work correctly", () => {
  testFlowContext({}).ui.display.grid({
    data: [{ name: "John", age: 30 }],
    render: (data) => {
      expectTypeOf(data).toHaveProperty("name").toBeString();
      expectTypeOf(data).toHaveProperty("age").toBeNumber();
      return {
        title: data.name,
        description: `Age: ${data.age}`,
      };
    },
  });
});
