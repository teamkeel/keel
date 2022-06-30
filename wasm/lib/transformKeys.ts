import deepMapKeys from 'deep-map-keys'

function snakeToCamel(key: string) {
  return key.replace(/_(\w)/g, (match, char) => char.toUpperCase());
}

export default (obj: Record<string, any>) : Record<string, any> => {
  return deepMapKeys(obj, (_, key) => snakeToCamel(key))
}
