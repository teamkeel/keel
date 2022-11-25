import { Config } from "../types";
import handle from ".";

test("when the custom function returns expected value", async () => {
  const config: Config = {
    functions: {
      createPost: () => {
        return {
          title: "a post",
          id: "abcde",
        };
      },
    },
    api: {},
  };

  expect(await handle("/createPost", { title: "a post" }, config)).toEqual({
    title: "a post",
    id: "abcde",
  });
});

test("when the custom function doesnt return a value", async () => {
  const config: Config = {
    functions: {
      createPost: () => {},
    },
    api: {},
  };
  await expect(
    handle("/createPost", { title: "a post" }, config)
  ).rejects.toThrowError("no result returned from custom function");
});
