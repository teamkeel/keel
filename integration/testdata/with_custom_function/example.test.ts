import { test, expect, Actions, Person, logger } from '@teamkeel/testing'

test('creating a person', async () => {
  const { object } = await Actions.createPerson({ name: 'foo', gender: 'female', nINumber: '282' })
  expect.equal(object.name, 'foo')
})

test('fetching a person by id', async () => {
  const { object: person } = await Person.create({ name: 'bar', gender: 'male', nINumber: '123' })
  const { object } = await Actions.getPerson({ id: person.id })

  expect.equal(object.id, person.id)
  expect.equal(object.name, person.name)
})

test('fetching person by additional unique field (not PK)', async () => {
  const { object: person } = await Person.create({ name: 'bar', gender: 'male', nINumber: '333' })

  const { object } = await Actions.getPerson({ nINumber: '333' })

  expect.equal(object.id, person.id)
})

test('listing', async () => {
  await Person.create({ name: 'fred', gender: 'male', nINumber: '000' })
  const { object: x11 } = await Person.create({ name: 'X11', gender: 'alien', nINumber: '920' })
  const { object: x22 } =  await Person.create({ name: 'X22', gender: 'alien', nINumber: '902' })

  const { collection: aliens } = await Actions.listPeople({ gender: 'alien' })

  const alienNames = aliens.map((a) => a.name)

  expect.equal(alienNames, [x11.name, x22.name])
})

test('deletion', async () => {
  const { object: person } = await Person.create({ name: 'fred', gender: 'male', nINumber: '678' })

  const { success } = await Actions.deletePerson({ id: person.id })

  expect.equal(success, true)
})

test('updating', async () => {
  const { object: person } = await Person.create({ name: 'fred', gender: 'male', nINumber: '678' })

  const { object: updatedPerson } = await Actions.updatePerson({ where: { id: person.id }, values: { name: 'paul' }})

  expect.equal(updatedPerson.name, 'paul')
  expect.equal(updatedPerson.id, person.id)
})
