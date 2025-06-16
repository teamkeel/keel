function getFunction({ model, whereInputs }) {
  return function (hooks = {}) {
    return async function (ctx, inputs) {
      let wheres = {};
      for (const key of whereInputs) {
        if (key in inputs) {
          wheres[key] = inputs[key];
        }
      }

      let query = model.where(wheres);

      if (hooks.beforeQuery) {
        query = await runtime.tracing.withSpan("beforeQuery", () => {
          return hooks.beforeQuery(ctx, inputs, query);
        });
      }

      const constructor = query?.constructor?.name;
      if (constructor === "QueryBuilder") {
        query = await query.findOne();
      }

      if (query === null) {
        return null;
      }

      if (hooks.afterQuery) {
        query = await runtime.tracing.withSpan("afterQuery", () => {
          return hooks.afterQuery(ctx, inputs, query);
        });
      }

      return query;
    };
  };
}
