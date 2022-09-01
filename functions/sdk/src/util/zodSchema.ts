import { z } from "zod";
import { ModelDefinition, ScalarTypes } from "../types";

export const buildZodSchemaFromModelDefinition = <T>(
  definition: ModelDefinition<T>
) => {
  return z.object(
    Object.entries(definition).reduce((acc, [key, value]) => {
      switch (value as ScalarTypes) {
        case "boolean":
          acc[key] = z.boolean();
          break;
        case "string":
          acc[key] = z.string();
          break;
        case "number":
          acc[key] = z.number();
          break;
        case "date":
          acc[key] = z.date();
        default:
          acc[key] = z.unknown();
      }
      return acc;
    }, {})
  );
};
