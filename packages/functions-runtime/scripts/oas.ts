import * as TJS from "typescript-json-schema";
import * as fs from "fs";
import * as path from "path";
import { fileURLToPath } from "url";
import { dirname } from "path";

// ************************************************************
// Generate the json schema from the UiApiUiConfig type
// ************************************************************

// @ts-ignore
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const program = TJS.getProgramFromFiles(
  [path.resolve("./src/flows/ui/index.ts")],
  {
    strictNullChecks: true,
    esModuleInterop: true,
    skipLibCheck: true,
    target: "ES2020",
    module: "ESNext",
    moduleResolution: "node",
    lib: ["ES2020", "DOM"],
  }
);

const generator = TJS.buildGenerator(program, {
  required: true,
  noExtraProps: true,
});

if (!generator) {
  throw new Error("Failed to create generator");
}

try {
  // Generate the schema just for our entry type
  const schema = generator.getSchemaForSymbol("UiApiUiConfig");

  const outputDir = path.resolve(__dirname, "../../../runtime/openapi");
  if (!fs.existsSync(outputDir)) {
    console.error("Output directory does not exist");
    process.exit(1);
  }

  // Write to file
  fs.writeFileSync(
    path.resolve(outputDir, "uiConfig.json"),
    JSON.stringify(schema, null, 2)
  );

  console.log("Successfully generated OpenAPI schema");
} catch (error) {
  console.error("Error generating schema:", error);
  process.exit(1);
}
