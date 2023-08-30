const { withSpan } = require("./tracing");
const { PROTO_ACTION_TYPES } = require("./consts");

const HOOK_TYPES = {
  BEFORE_WRITE: "beforeWrite",
  AFTER_WRITE: "afterWrite",
  BEFORE_QUERY: "beforeQuery",
  AFTER_QUERY: "afterQuery",
};

// **
// * hookDefined will first check if the hook type is applicable to the action type (some action types don't support some hooks)
// * then it will check if the user defined the hook in the hooks object passed into this function
// **
const hookDefined = (hooks, hookType) => {
  return hooks[hookType] !== undefined;
};

const PERMITTED_ACTION_TYPES = [
  PROTO_ACTION_TYPES.CREATE,
  PROTO_ACTION_TYPES.DELETE,
  PROTO_ACTION_TYPES.GET,
  PROTO_ACTION_TYPES.LIST,
  PROTO_ACTION_TYPES.UPDATE,
];

// **
// * defaultImplementation provides the default implementation of an
// * action type, interspersing with calls to user defined hook functions
// * @param hooks - the hooks object tha user passed to the function
// * @param action - the action object, containing the name, type and modelAPI for the action
// * @returns an async function that provides the base implementation for each action type.
const defaultImplementation = (hooks = {}, action) => {
  const { name: actionName, type: actionType, modelAPI } = action;

  if (!PERMITTED_ACTION_TYPES.includes(actionType)) {
    throw new Error("unsupported action type " + actionType);
  }

  return async function (ctx, inputs) {
    return await withSpan(
      `${actionName}.defaultImplementation`,
      async (span) => {
        let values, wheres, data;

        switch (actionType) {
          case PROTO_ACTION_TYPES.CREATE:
            // the values are the whole of the inputs
            // for a create action
            values = Object.assign({}, inputs);

            if (hookDefined(hooks, HOOK_TYPES.BEFORE_WRITE)) {
              await withSpan(`${actionName}.beforeWrite`, async (span) => {
                values = await hooks.beforeWrite(
                  ctx,
                  deepFreeze(inputs),
                  values
                );
              });
            }

            data = await modelAPI.create(values);

            if (hookDefined(hooks, HOOK_TYPES.AFTER_WRITE)) {
              await withSpan(`${actionName}.afterWrite`, async (span) => {
                await hooks.afterWrite(ctx, deepFreeze(inputs), data);
              });
            }

            return data;
          case PROTO_ACTION_TYPES.UPDATE:
            values = Object.assign({}, inputs.values);
            wheres = Object.assign({}, inputs.where);

            // call beforeWrite hook (if defined)
            if (hookDefined(hooks, HOOK_TYPES.BEFORE_WRITE)) {
              await withSpan(`${actionName}.beforeWrite`, async (span) => {
                values = await hooks.beforeWrite(
                  ctx,
                  deepFreeze(inputs),
                  values
                );
              });
            }

            if (hookDefined(hooks, HOOK_TYPES.BEFORE_QUERY)) {
              await withSpan(`${actionName}.beforeQuery`, async (span) => {
                data = await hooks.beforeQuery(ctx, deepFreeze(inputs), values);
              });
            } else {
              // when no beforeQuery hook is defined, use the default implementation
              data = await modelAPI.update(wheres, values);
            }

            // call afterQuery hook (if defined)
            if (hookDefined(hooks, HOOK_TYPES.AFTER_QUERY)) {
              await withSpan(`${actionName}.afterQuery`, async (span) => {
                data = await hooks.afterQuery(ctx, deepFreeze(inputs), data);
              });
            }

            // call afterWrite hook (if defined)
            if (hookDefined(hooks, HOOK_TYPES.AFTER_WRITE)) {
              await withSpan(`${actionName}.afterWrite`, async (span) => {
                await hooks.afterWrite(ctx, deepFreeze(inputs), data);
              });
            }

            return data;
          case PROTO_ACTION_TYPES.DELETE:
            wheres = Object.assign({}, inputs);

            if (hookDefined(hooks, HOOK_TYPES.BEFORE_QUERY)) {
              let builder = modelAPI.where(wheres);

              // we don't know if the return value of the hook is a Promise<string> or a {Model}QueryBuilder
              // so we await the result and check the constructor.name value to determine what type of
              // return value we are dealing with
              let resolvedValue;

              await withSpan(`${actionName}.beforeQuery`, async (span) => {
                resolvedValue = await hooks.beforeQuery(
                  ctx,
                  deepFreeze(inputs),
                  builder
                );
              });

              if (usesQueryBuilder(resolvedValue)) {
                builder = resolvedValue;

                // call .delete() on the query builder instance, which will contain the base constraints
                // calculated from the inputs.where values, as well as any additional constraints added
                // by the user in the beforeQuery hook

                span.addEvent(builder.sql());

                data = builder.delete();
              } else {
                // in this case we know that the user has defined a beforeQuery hook
                // that returns a Promise<string> where string is the deleted id.
                data = resolvedValue;
              }
            } else {
              // provide a default implementation in the case where no beforeQuery
              // hook has been defined by the user
              data = await modelAPI.delete(wheres);
            }

            if (hookDefined(hooks, HOOK_TYPES.AFTER_QUERY)) {
              await withSpan(`${actionName}.afterQuery`, async (span) => {
                data = await hooks.afterQuery(ctx, deepFreeze(inputs), data);
              });
            }

            return data;
          case PROTO_ACTION_TYPES.GET:
            wheres = Object.assign({}, inputs.where);

            if (hookDefined(hooks, HOOK_TYPES.BEFORE_QUERY)) {
              let builder = modelAPI.where(wheres);

              // we don't know if the return value of the hook is a Promise<string> or a {Model}QueryBuilder
              // so we await the result and check the constructor.name value to determine what type of
              // return value we are dealing with
              let resolvedValue;

              await withSpan(`${actionName}.beforeQuery`, async (span) => {
                resolvedValue = await hooks.beforeQuery(
                  ctx,
                  deepFreeze(inputs),
                  builder
                );
              });

              if (usesQueryBuilder(resolvedValue)) {
                builder = resolvedValue;

                // call .delete() on the query builder instance, which will contain the base constraints
                // calculated from the inputs.where values, as well as any additional constraints added
                // by the user in the beforeQuery hook

                span.addEvent(builder.sql());

                data = builder.findOne();
              } else {
                // in this case we know that the user has defined a beforeQuery hook
                // that returns a Promise<string> where string is the deleted id.
                data = resolvedValue;
              }
            } else {
              // provide a default implementation in the case where no beforeQuery
              // hook has been defined by the user
              data = await modelAPI.findOne(wheres);
            }

            if (hookDefined(hooks, HOOK_TYPES.AFTER_QUERY)) {
              await withSpan(`${actionName}.afterQuery`, async (span) => {
                data = await hooks.afterQuery(ctx, deepFreeze(inputs), data);
              });
            }

            return data;
          case PROTO_ACTION_TYPES.LIST:
            wheres = Object.assign({}, inputs.where);

            let builder = modelAPI.where(wheres);

            if (hookDefined(hooks, HOOK_TYPES.BEFORE_QUERY)) {
              // we don't know if the return value of the hook is a Promise<string> or a {Model}QueryBuilder
              // so we await the result and check the constructor.name value to determine what type of
              // return value we are dealing with
              let resolvedValue;

              await withSpan(`${actionName}.beforeQuery`, async (span) => {
                resolvedValue = await hooks.beforeQuery(
                  ctx,
                  deepFreeze(inputs),
                  builder
                );
              });

              if (usesQueryBuilder(resolvedValue)) {
                builder = resolvedValue;

                // call .delete() on the query builder instance, which will contain the base constraints
                // calculated from the inputs.where values, as well as any additional constraints added
                // by the user in the beforeQuery hook

                span.addEvent(builder.sql());

                data = builder.findMany();
              } else {
                // in this case we know that the user has defined a beforeQuery hook
                // that returns a Promise<string> where string is the deleted id.
                data = resolvedValue;
              }
            } else {
              // provide a default implementation in the case where no beforeQuery
              // hook has been defined by the user
              data = await builder.findMany();
            }

            if (hookDefined(hooks, HOOK_TYPES.AFTER_QUERY)) {
              await withSpan(`${actionName}.afterQuery`, async (span) => {
                data = await hooks.afterQuery(ctx, deepFreeze(inputs), data);
              });
            }

            return data;
          default:
            throw new Error("unhandled action type " + actionType);
        }
      }
    );
  };
};

const usesQueryBuilder = (resolvedValue) => {
  const constructor = resolvedValue?.constructor?.name;

  return constructor === "QueryBuilder";
};

const deepFreeze = (o) => {
  if (o === null || typeof o !== "object") return o;
  return new Proxy(o, {
    get(obj, prop) {
      return deepFreeze(obj[prop]);
    },
    set(obj, prop) {
      throw new Error(
        "Input " +
          JSON.stringify(obj) +
          " cannot be modified. Did you mean to modify values instead?"
      );
    },
  });
};

module.exports.defaultImplementation = defaultImplementation;
