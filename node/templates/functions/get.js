function getFunction({ model, whereInputs }) {
  return function (hooks = {}) {
    return async function (ctx, inputs) {
      let wheres = {};
      for (const key of whereInputs) {
        if (key in inputs) {
          wheres[key] = inputs[key];
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
        data = await data.findOne();
      }

      if (data === null) {
        return null;
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
