declare module '@teamkeel/wasm/build' {
    export {};
}

declare module '@teamkeel/wasm/index' {
    interface ValidateOptions {
        color: boolean;
    }

    interface KeelAPI {
        format: (schemaString: string) => string;
        validate: (schemaString: string, options?: ValidateOptions) => string;
    }
    const instantiate: () => Promise<KeelAPI>;
    export default instantiate;

}

declare module '@teamkeel/wasm/lib/wasm_exec' {

}

declare module '@teamkeel/wasm/lib/wasm_exec_node' {
    export {};
}

declare module '@teamkeel/wasm' {
    import main = require('@teamkeel/wasm/index');
    export = main;
}