# IAM SessionIdentity demo

I've created this repo to gain understanding on how the `SessionIdentity` works in the context of `sts:AssumeRole` operation.

## Deployment

- Install necessary Go dependencies

  ```sh
  cd get-item
  go install
  ```

- Build the app

  ```sh
  sam build
  ```

- Deploy the app

  ```sh
  sam deploy --guided
  ```

## Learnings

- Remember that to set the `SessionIdentity` the trust policy of the role has to allow for it
- The `SessionIdentity` is sticky and this stickiness is validated by IAM.

  - **If your role already has `SessionIdentity` set, you will not be able to set it again while assuming a role**.
  - The error that you get from IAM is pretty verbose. This is in contrast to, for example, S3 where 403 will be returned if the object that you are trying to retrieve does not exist.
