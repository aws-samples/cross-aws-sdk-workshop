import * as iam from 'monocdk/aws-iam';

interface PolicyDocumentProps {
  readonly statements?: iam.PolicyStatement[];
}

/**
 * Helper to create an IAM policy document inline without the need to separate
 * the creation of the statement into multiple lines, and local variables.
 *
 * Can be used to help making IAM Policies or Roles inline.
 */
export function makePolicyDocument(
  props: PolicyDocumentProps
): iam.PolicyDocument {
  let document = new iam.PolicyDocument({});

  if (props.statements) {
    document.addStatements(...props.statements);
  }

  return document;
}

interface PolicyStatementProps {
  readonly effect: iam.Effect;
  readonly actions?: string[];
  readonly resources?: string[];
  readonly principals?: iam.IPrincipal[];
}

/**
 * Helper to create an IAM policy statement inline without the need to separate
 * the creation of the statement into multiple lines, and local variables.
 *
 * Can be used to help making IAM Policies or Roles inline.
 */
export function makePolicyStatement(
  props: PolicyStatementProps
): iam.PolicyStatement {
  let statement = new iam.PolicyStatement({ effect: props.effect });

  if (props.actions) {
    statement.addActions(...props.actions);
  }
  if (props.resources) {
    statement.addResources(...props.resources);
  }
  if (props.principals) {
    statement.addPrincipals(...props.principals);
  }

  return statement;
}
