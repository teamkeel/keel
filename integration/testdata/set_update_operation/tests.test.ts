import { test, expect, actions, Thing } from '@teamkeel/testing'

// Outstanding tests & features:
// - setting using implicit inputs
// - setting using explicit inputs
// - setting using explicit inputs overriding implicit write

/* 
  Text Type 
*/

test('text set attribute - set to goodbye - is goodbye', async () => {
  const { object: thing } = await actions.create({})
  await actions.updateText({ where: { id: thing.id } })
  const { object: updated } = await actions.get({ id: thing.id })
  expect(updated.text).toEqual("goodbye")
})

test('text set attribute - set to null - is null', async () => {
  const { object: thing } = await actions.create({})
  await actions.updateNullText({ where: { id: thing.id } })
  const { object: updated } = await actions.get({ id: thing.id })
  expect(updated.text).toEqual(null)
})

/* 
  Number Type 
*/

test('number set attribute - set to 5 - is 5', async () => {
  const { object: thing } = await actions.create({})
  await actions.updateNumber({ where: { id: thing.id } })
  const { object: updated } = await actions.get({ id: thing.id })
  expect(updated.number).toEqual(5)
})

test('number set attribute - set to null - is null', async () => {
  const { object: thing } = await actions.create({})
  await actions.updateNullNumber({ where: { id: thing.id } })
  const { object: updated } = await actions.get({ id: thing.id })
  expect(updated.number).toEqual(null)
})

/* 
  Boolean Type 
*/

test('boolean set attribute - set to true - is true', async () => {
  const { object: thing } = await actions.create({})
  await actions.updateBoolean({ where: { id: thing.id } })
  const { object: updated } = await actions.get({ id: thing.id })
  expect(updated.boolean).toEqual(true)
})

test('boolean set attribute - set to null - is null', async () => {
  const { object: thing } = await actions.create({})
  await actions.updateNullBoolean({ where: { id: thing.id } })
  const { object: updated } = await actions.get({ id: thing.id })
  expect(updated.boolean).toEqual(null)
})

/* 
  Enum Type 
*/

test('enum set attribute - set to TypeTwo - is TypeTwo', async () => {
  const { object: thing } = await actions.create({})
  await actions.updateEnum({ where: { id: thing.id } })
  const { object: updated } = await actions.get({ id: thing.id })
  expect(updated.enum).toEqual("TypeTwo")
})

test('enum set attribute - set to null - is null', async () => {
  const { object: thing } = await actions.create({})
  await actions.updateNullEnum({ where: { id: thing.id } })
  const { object: updated } = await actions.get({ id: thing.id })
  expect(updated.enum).toEqual(null)
})