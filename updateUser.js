import { DynamoDBClient } from "@aws-sdk/client-dynamodb";
import { DynamoDBDocumentClient, UpdateCommand } from "@aws-sdk/lib-dynamodb";

const client = new DynamoDBClient();
const ddbDocClient = DynamoDBDocumentClient.from(client);

export const handler = async (event) => {
  const { id, stripe_customer_id, payment_link, username, password, payment_status } = event.arguments;

  const params = {
    TableName: 'User',
    Key: { id },
    UpdateExpression: 'set stripe_customer_id = :sc, payment_link = :pl, username = :un, password = :pw, payment_status = :ps',
    ExpressionAttributeValues: {
      ':sc': stripe_customer_id,
      ':pl': payment_link,
      ':un': username,
      ':pw': password,
      ':ps': payment_status
    },
    ReturnValues: 'ALL_NEW'
  };

  const data = await ddbDocClient.send(new UpdateCommand(params));

  return data.Attributes;
};
