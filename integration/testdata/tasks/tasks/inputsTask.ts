import { InputsTask, Priority } from "@teamkeel/sdk";

export default InputsTask({}, async (ctx, inputs) => {
    // Type assertions - these will fail at runtime if types are wrong
    if (typeof inputs.textField !== "string") {
        throw new Error(`textField should be string, got ${typeof inputs.textField}`);
    }
    if (typeof inputs.numberField !== "number") {
        throw new Error(`numberField should be number, got ${typeof inputs.numberField}`);
    }
    if (typeof inputs.booleanField !== "boolean") {
        throw new Error(`booleanField should be boolean, got ${typeof inputs.booleanField}`);
    }
    // Date fields come as Date objects from the runtime
    if (!(inputs.dateField instanceof Date)) {
        throw new Error(`dateField should be Date, got ${typeof inputs.dateField}: ${inputs.dateField}`);
    }
    if (!(inputs.timestampField instanceof Date)) {
        throw new Error(`timestampField should be Date, got ${typeof inputs.timestampField}: ${inputs.timestampField}`);
    }
    if (typeof inputs.decimalField !== "number") {
        throw new Error(`decimalField should be number, got ${typeof inputs.decimalField}`);
    }
    if (!Object.values(Priority).includes(inputs.enumField)) {
        throw new Error(`enumField should be Priority enum value, got ${inputs.enumField}`);
    }
    // Optional field can be null or string
    if (inputs.optionalTextField !== null && typeof inputs.optionalTextField !== "string") {
        throw new Error(`optionalTextField should be string or null, got ${typeof inputs.optionalTextField}`);
    }

    // Value assertions - verify actual values match expected test data
    // Only run for the "typed inputs" test (identified by textField value)
    if (inputs.textField === "hello world") {
        if (inputs.numberField !== 42) {
            throw new Error(`numberField value should be 42, got ${inputs.numberField}`);
        }
        if (inputs.booleanField !== true) {
            throw new Error(`booleanField value should be true, got ${inputs.booleanField}`);
        }
        if (inputs.dateField.toISOString() !== "2025-07-15T00:00:00.000Z") {
            throw new Error(`dateField value should be "2025-07-15T00:00:00.000Z", got "${inputs.dateField.toISOString()}"`);
        }
        if (inputs.timestampField.toISOString() !== "2025-07-15T14:30:00.000Z") {
            throw new Error(`timestampField value should be "2025-07-15T14:30:00.000Z", got "${inputs.timestampField.toISOString()}"`);
        }
        if (inputs.decimalField !== 123.456) {
            throw new Error(`decimalField value should be 123.456, got ${inputs.decimalField}`);
        }
        if (inputs.enumField !== Priority.High) {
            throw new Error(`enumField value should be Priority.High, got ${inputs.enumField}`);
        }
        if (inputs.optionalTextField !== "optional value") {
            throw new Error(`optionalTextField value should be "optional value", got "${inputs.optionalTextField}"`);
        }
    }

    // Value assertions for the "optional fields" test
    if (inputs.textField === "test") {
        if (inputs.numberField !== 1) {
            throw new Error(`numberField value should be 1, got ${inputs.numberField}`);
        }
        if (inputs.booleanField !== false) {
            throw new Error(`booleanField value should be false, got ${inputs.booleanField}`);
        }
        if (inputs.dateField.toISOString() !== "2025-01-01T00:00:00.000Z") {
            throw new Error(`dateField value should be "2025-01-01T00:00:00.000Z", got "${inputs.dateField.toISOString()}"`);
        }
        if (inputs.timestampField.toISOString() !== "2025-01-01T00:00:00.000Z") {
            throw new Error(`timestampField value should be "2025-01-01T00:00:00.000Z", got "${inputs.timestampField.toISOString()}"`);
        }
        if (inputs.decimalField !== 0.5) {
            throw new Error(`decimalField value should be 0.5, got ${inputs.decimalField}`);
        }
        if (inputs.enumField !== Priority.Low) {
            throw new Error(`enumField value should be Priority.Low, got ${inputs.enumField}`);
        }
        if (inputs.optionalTextField !== null) {
            throw new Error(`optionalTextField value should be null, got "${inputs.optionalTextField}"`);
        }
    }

    return ctx.complete({
        data: {
            textField: inputs.textField,
            numberField: inputs.numberField,
            booleanField: inputs.booleanField,
            dateField: inputs.dateField.toISOString(),
            timestampField: inputs.timestampField.toISOString(),
            decimalField: inputs.decimalField,
            enumField: inputs.enumField,
            optionalTextField: inputs.optionalTextField,
        }
    });
});
