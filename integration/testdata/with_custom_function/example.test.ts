import { test, expect, actions, Person } from '@teamkeel/testing'

test('creating a person', async () => {
  const { object } = await actions
    .createPerson({ name: 'foo', gender: 'female', nINumber: '282' })

  expect(object.name).toEqual('foo')
})

test('fetching a person by id', async () => {
  const { object: person } = await Person.create({ name: 'bar', gender: 'male', nINumber: '123' })
  const { object } = await actions
    .getPerson({ id: person.id })

  expect(object.id).toEqual(person.id)
  expect(object.name).toEqual(person.name)
})

test('fetching person by additional unique field (not PK)', async () => {
  const { object: person } = await Person.create({ name: 'bar', gender: 'male', nINumber: '333' })

  const { object } = await actions
    .getPersonByNINumber({ nINumber: '333' })

  expect(object.id).toEqual(person.id)
})

test('listing', async () => {
  await Person.create({ name: 'fred', gender: 'male', nINumber: '000' })
  const { object: x11 } = await Person.create({ name: 'X11', gender: 'alien', nINumber: '920' })
  const { object: x22 } =  await Person.create({ name: 'X22', gender: 'alien', nINumber: '902' })

  const { collection: aliens } = await actions
    .listPeople({ gender: { equals: 'alien' } })

  const alienNames = aliens.map((a) => a.name)

  expect(alienNames).toEqual([x11.name, x22.name])
})

test('deletion', async () => {
  const { object: person } = await Person.create({ name: 'fred', gender: 'male', nINumber: '678' })

  const { success } = await actions
    .deletePerson({ id: person.id })

  expect(success).toEqual(true)
})

test('updating', async () => {
  const { object: person } = await Person.create({ name: 'fred', gender: 'male', nINumber: '678' })

  const { object: updatedPerson } = await actions
    .updatePerson({ where: { id: person.id }, values: { name: 'paul' }})

  expect(updatedPerson.name).toEqual('paul')
  expect(updatedPerson.id).toEqual(person.id)
})
