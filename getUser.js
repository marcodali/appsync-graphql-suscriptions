import { DynamoDBClient } from "@aws-sdk/client-dynamodb";
import { DynamoDBDocumentClient, GetCommand } from "@aws-sdk/lib-dynamodb";

const client = new DynamoDBClient();
const ddbDocClient = DynamoDBDocumentClient.from(client);

export const handler = async (event) => {
  const { id } = event.arguments;

  const params = {
    TableName: 'User',
    Key: { id }
  };

  const data = await ddbDocClient.send(new GetCommand(params));

  return data.Item;
};
