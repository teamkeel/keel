function createFunction({ model, valueInputs }) {
  return function (hooks = {}) {
    return async function (ctx, inputs) {
      let values = {};
      for (const key of valueInputs) {
        if (key in inputs) {
          values[key] = inputs[key];
        }
      }

      if (hooks.beforeWrite) {
        values = await runtime.tracing.withSpan("beforeWrite", () => {
          return hooks.beforeWrite(ctx, inputs, values);
        });
      }

      let data = await model.create(values);

      if (hooks.afterWrite) {
        const v = await runtime.tracing.withSpan("afterWrite", () => {
          return hooks.afterWrite(ctx, inputs, data);
        });
        if (v !== undefined) {
          data = v;
        }
      }

      return data;
    };
  };
}
