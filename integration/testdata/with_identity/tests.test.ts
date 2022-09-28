import { test, expect, actions, Post } from '@teamkeel/testing'

test('authorization successful', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')
    .createPost({title: 'temp'});

  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')
    .getPost({ id: post.id });

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, false)
})

test('authorization failed', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')
    .createPost({title: 'temp'});

  const { errors } = await actions
    .withIdentity('2FQqdDYm47mEjgEGsUTtVYbDmuM')
    .getPost({ id: post.id });

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

function hasAuthorizationError(errors?): boolean {
  if (errors == null)
    return false;

  var hasError = false
   errors.forEach(function(error) {
    if(error.message == 'not authorized to access this operation') {
      hasError = true
    }
  });
  
  return hasError;
}