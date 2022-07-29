import { BaseAPI, Model } from '../types'
import { buildModelApis } from './model'

export default (models: Model[]) : BaseAPI => {
  const api : BaseAPI = {
    models: buildModelApis(models)
  }

  return api
}
