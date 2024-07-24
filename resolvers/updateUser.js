import { util } from '@aws-appsync/utils';

/**
 * Updates an item in the DynamoDB table.
 * @param {import('@aws-appsync/utils').Context} ctx the context
 * @returns {import('@aws-appsync/utils').DynamoDBUpdateItemRequest} the request
 */
export function request(ctx) {
    const updateExpressions = [];
    const expressionAttributeValues = {};
    const expressionAttributeNames = {};

    // Iterate over all arguments to dynamically build the update expression
    for (const key in ctx.args) {
        if (key !== 'id' && ctx.args[key] !== null && ctx.args[key] !== undefined) {
            updateExpressions.push(`#${key} = :${key}`);
            expressionAttributeValues[`:${key}`] = util.dynamodb.toDynamoDB(ctx.args[key]);
            expressionAttributeNames[`#${key}`] = key;
        }
    }

    return {
        operation: 'UpdateItem',
        key: util.dynamodb.toMapValues({ id: ctx.args.id }),
        update: {
            expression: `SET ${updateExpressions.join(', ')}`,
            expressionValues: expressionAttributeValues,
            expressionNames: expressionAttributeNames,
        }
    };
}

/**
 * Returns the updated item or throws an error if the operation failed.
 * @param {import('@aws-appsync/utils').Context} ctx the context
 * @returns {*} the updated item
 */
export function response(ctx) {
    if (ctx.error) {
        util.error(ctx.error.message, ctx.error.type);
    }
    return ctx.result;
}
