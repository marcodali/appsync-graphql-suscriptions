import { DynamoDBClient } from "@aws-sdk/client-dynamodb";
import { DynamoDBDocumentClient, PutCommand } from "@aws-sdk/lib-dynamodb";

const client = new DynamoDBClient();
const ddbDocClient = DynamoDBDocumentClient.from(client);

export const handler = async (event) => {
  const { stripe_customer_id, payment_link, username, password, payment_status } = event.arguments;

  const params = {
    TableName: 'User',
    Item: {
      id: crypto.randomUUID(),
      stripe_customer_id,
      payment_link,
      username,
      password,
      payment_status
    }
  };

  await ddbDocClient.send(new PutCommand(params));

  return params.Item;
};
