import * as TJS from "typescript-json-schema";
import * as fs from "fs";
import * as path from "path";
import { fileURLToPath } from "url";
import { dirname } from "path";
import alterschema from "alterschema";

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
  ref: true,
});

const generatorNoRef = TJS.buildGenerator(program, {
  required: true,
  noExtraProps: true,
  ref: false,
});

if (!generator || !generatorNoRef) {
  throw new Error("Failed to create generator");
}

try {
  const outputDir = path.resolve(__dirname, "../../../runtime/openapi");
  if (!fs.existsSync(outputDir)) {
    console.error("Output directory does not exist");
    process.exit(1);
  }

  // Generate the schema for ui config
  const schema = generator.getSchemaForSymbol("UiApiUiConfig");

  // Convert to latest JSON schema spec
  // @ts-ignore
  const schemaConverted = await alterschema(schema, "draft7", "2020-12");

  // Write to file
  fs.writeFileSync(
    path.resolve(outputDir, "uiConfig.json"),
    JSON.stringify(schemaConverted, null, 2)
  );

  // Generate the schema for ui config
  const flowSchema = generatorNoRef.getSchemaForSymbol("FlowConfig");

  // Convert to latest JSON schema spec
  // @ts-ignore
  const flowSchemaConverted = await alterschema(
    flowSchema,
    "draft7",
    "2020-12"
  );

  // Write to file
  fs.writeFileSync(
    path.resolve(outputDir, "flowConfig.json"),
    JSON.stringify(flowSchemaConverted, null, 2)
  );

  console.log("Successfully generated OpenAPI schema");
} catch (error) {
  console.error("Error generating schema:", error);
  process.exit(1);
}
