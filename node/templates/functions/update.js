function updateFunction({ model, whereInputs, valueInputs }) {
  return function (hooks = {}) {
    return async function (ctx, inputs) {
      let wheres = {};
      let values = {};
      for (const key of whereInputs) {
        if (inputs.where && key in inputs.where) {
          wheres[key] = inputs.where[key];
        }
      }
      for (const key of valueInputs) {
        if (inputs.values && key in inputs.values) {
          values[key] = inputs.values[key];
        }
      }

      let data = model.where(wheres);

      if (hooks.beforeQuery) {
        data = await runtime.tracing.withSpan("beforeQuery", () => {
          if (!inputs || Object.keys(inputs).length === 0) {
            return hooks.beforeQuery(ctx, data);
          } else {
            return hooks.beforeQuery(ctx, inputs, data);
          }
        });
      }

      const constructor = data?.constructor?.name;
      if (constructor === "QueryBuilder") {
        data = await data.findOne();
      }

      if (data === null) {
        throw new NoResultError();
      }

      if (hooks.beforeWrite) {
        values = await runtime.tracing.withSpan("beforeWrite", () => {
          if (!inputs || Object.keys(inputs).length === 0) {
            return hooks.beforeWrite(ctx, values, data);
          } else {
            return hooks.beforeWrite(ctx, inputs, values, data);
          }
        });
      }

      data = await model.update({ id: data.id }, values);

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
