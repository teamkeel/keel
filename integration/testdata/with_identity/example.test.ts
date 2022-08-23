import { test, expect, Actions, Customer, logger } from '@teamkeel/testing'

test('it passes', async () => {
  const customer = await Customer.create({
    name: 'Adam'
  })
  logger.log('here')
  const result = await Actions
                        .withIdentity(customer.identity)
                        .createOrder({ title: 'foo' })

  logger.log(result)
  expect.equal(result.title, 'foo')
})
