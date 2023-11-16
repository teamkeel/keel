function listFunction({ model, whereInputs }) {
  return function (hooks = {}) {
    return async function (ctx, inputs) {
      let wheres = {};
      for (const key of whereInputs) {
        if (inputs.where && key in inputs.where) {
          wheres[key] = inputs.where[key];
        }
      }

      let data = model.where(wheres);

      if (hooks.beforeQuery) {
        data = await runtime.tracing.withSpan("beforeQuery", () => {
          return hooks.beforeQuery(ctx, inputs, data);
        });
      }

      const constructor = data?.constructor?.name;
      if (constructor === "QueryBuilder") {
        data = await data.findMany({ limit: inputs.first });
      }

      if (hooks.afterQuery) {
        data = await runtime.tracing.withSpan("afterQuery", () => {
          return hooks.afterQuery(ctx, inputs, data);
        });
      }

      return data;
    };
  };
}
