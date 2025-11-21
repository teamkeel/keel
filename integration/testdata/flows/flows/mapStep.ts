import { MapStep, models } from "@teamkeel/sdk";

export default MapStep(
    {
        title: "Map step",
    },
    async (ctx) => {
        const mapResult = await ctx.step("create map", async () => {
            const myMap = new Map<string, any>();
            myMap.set("name", "Keelson");
            myMap.set("age", 25);
            myMap.set("active", true);
            myMap.set("nested", { city: "London", country: "UK" });

            return myMap;
        });

        // Verify the returned value - Maps are converted to plain objects during serialization
        await ctx.step("verify map object", async () => {
            // Cast to any since the type system doesn't know the exact shape
            const obj = mapResult as any;

            return {
                // Check that the values are accessible as object properties
                hasName: obj.name === "Keelson",
                hasAge: obj.age === 25,
                hasActive: obj.active === true,
                hasNested: obj.nested?.city === "London" && obj.nested?.country === "UK",
                // Maps are serialized to objects, so this will be false
                isMap: obj instanceof Map,
                isObject: typeof obj === "object" && obj !== null,
            };
        });
    }
);

