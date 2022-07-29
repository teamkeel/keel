import { Model, ModelApi, asSchema, MapSchema } from '../types'

export const buildModelApis = (models: Model[]) : Record<string, unknown> =>  {
  const ret = {}

  models.forEach((model) => {
    const schema = asSchema(model.definition)
    type M = MapSchema<typeof schema>
    const api = buildModelApi<M>()

    ret[model.name] = api
  })

  return ret
}

export const buildModelApi = <T>() : ModelApi<T> => {
  return {
    create: async (inputs) => Promise.resolve({} as T),
    delete: async (id) => Promise.resolve(true),
    find: async (inputs) => Promise.resolve({} as T),
    update: async (id, inputs) => Promise.resolve({} as T),
    findMany: async (inputs) => Promise.resolve([])
  }
}
