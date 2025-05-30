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
          if (!inputs || Object.keys(inputs).length === 0) {
            return hooks.beforeWrite(ctx, values);
          } else {
            return hooks.beforeWrite(ctx, inputs, values);
          }
        });
      }

      let data = await model.create(values);

      if (hooks.afterWrite) {
        const v = await runtime.tracing.withSpan("afterWrite", () => {
          if (!inputs || Object.keys(inputs).length === 0) {
            return hooks.afterWrite(ctx, data);
          } else {
            return hooks.afterWrite(ctx, inputs, data);
          }
        });
        if (v !== undefined) {
          data = v;
        }
      }

      return data;
    };
  };
}
