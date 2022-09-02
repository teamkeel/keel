import { faker } from "@faker-js/faker";

import { ModelDefinition, ScalarTypes } from "./types";

const buildModel = <M extends Record<string, any>>(
  definition: ModelDefinition<M>
): M => {
  return Object.entries(definition).reduce((acc, [key, type]) => {
    switch (type as ScalarTypes) {
      case "boolean":
        acc[key] = faker.datatype.boolean();
        break;
      case "date":
        acc[key] = faker.date.past();
        break;
      case "number":
        acc[key] = faker.datatype.number();
        break;
      case "string":
        acc[key] = faker.random.words();
        break;
      default:
        throw new Error(
          `Seed data generation for ${type} (column ${key}) not yet supported`
        );
    }
    return acc;
  }, {} as Record<string, any>) as M;
};

export default buildModel;
