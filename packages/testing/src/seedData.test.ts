import { ModelDefinition } from "types";
import buildModel from "./seedData";

interface Post {
  title?: string;
  published?: boolean;
  rating?: number;
  createdAt?: Date;
}

describe("buildModel", () => {
  it("builds the object correctly", () => {
    const definition = {
      title: "string",
      published: "boolean",
      rating: "number",
      createdAt: "date",
    } as ModelDefinition<Post>;

    const instance = buildModel<Post>(definition);

    expect(instance.title).not.toBeUndefined();
    expect(instance.rating).not.toBeUndefined();
    expect(instance.createdAt).not.toBeUndefined();
    expect(instance.published).not.toBeUndefined();
  });
});
