import * as cdk from 'monocdk';
import * as ddb from 'monocdk/aws-dynamodb';
import * as iam from 'monocdk/aws-iam';
import * as lambda from 'monocdk/aws-lambda';
import * as lambda_nodejs from 'monocdk/aws-lambda-nodejs';
import * as s3 from 'monocdk/aws-s3';
import * as sfn from 'monocdk/aws-stepfunctions';
import { ApiGatewayFrontend } from './api-gateway';
import * as iamUtils from './iam-utils';
import { getNodejsBundlingOptions } from './lambda-utils';
import { TranscribeStateMachine } from './transcribe-statemachine';

export enum WorkshopLanguage {
  Go = 'go',
  Python = 'python',
  Java = 'java',
  Javascript = 'javascript',
}

export function parseWorkshopLanguage(language: string): WorkshopLanguage {
  switch (language) {
    case WorkshopLanguage.Go:
      return WorkshopLanguage.Go;

    case WorkshopLanguage.Python:
      return WorkshopLanguage.Python;

    case WorkshopLanguage.Java:
      return WorkshopLanguage.Java;

    case WorkshopLanguage.Javascript:
      return WorkshopLanguage.Javascript;

    default:
      throw new Error(`Unknown workshop language, ${language}`);
  }
}

export const ENV_KEY_PREFIX = 'AWS_SDK_WORKSHOP_';

const ENV_KEY_TRANSCRIBE_STATEMACHINE_ARN =
  ENV_KEY_PREFIX + 'TRANSCRIBE_STATEMACHINE_ARN';
const ENV_KEY_PODCAST_EPISODE_TABLE_NAME =
  ENV_KEY_PREFIX + 'PODCAST_EPISODE_TABLE_NAME';
const ENV_KEY_PODCAST_DATA_BUCKET_NAME =
  ENV_KEY_PREFIX + 'PODCAST_DATA_BUCKET_NAME';
const ENV_KEY_TRANSCRIBE_ACCESS_ROLE_ARN =
  ENV_KEY_PREFIX + 'TRANSCRIBE_ACCESS_ROLE_ARN';

const ENV_KEY_PODCAST_DATA_KEY_PREFIX =
  ENV_KEY_PREFIX + 'PODCAST_DATA_KEY_PREFIX';
const ENV_KEY_MAX_NUM_EPISODE_IMPORT =
  ENV_KEY_PREFIX + 'MAX_NUM_EPISODE_IMPORT';

const PODCAST_DATA_KEY_PREFIX = 'podcasts/';
const MAX_NUM_EPISODE_IMPORT = '5';

export interface CdkStackProps extends cdk.StackProps {
  workshopLanguage: WorkshopLanguage;
}

export class CdkStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props: CdkStackProps) {
    super(scope, id, props);

    const podcastBucket = new s3.Bucket(this, 'PodcastData', {
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
    });
    const transcribeAccessRole = makeTranscribeAccessBucketRole(
      this,
      'PodcastDataTranscribeAccessRole',
      podcastBucket
    );

    const podcastEpisodeTable = new ddb.Table(this, 'PodcastEpisode', {
      partitionKey: { type: ddb.AttributeType.STRING, name: 'id' },
    });

    const transcribeStateMachine = new TranscribeStateMachine(
      this,
      'TranscribePodcast',
      {
        ...makeTranscribeStatemachineLambdas(this, 'TranscribeHandler', {
          workshopLanguage: props.workshopLanguage,
          podcastBucket: podcastBucket,
          podcastEpisodeTable: podcastEpisodeTable,
          transcribeAccessRole: transcribeAccessRole,
        }),
      }
    ).stateMachine;

    const frontend = new ApiGatewayFrontend(this, 'PodcastApi', {
      ...makeApiEndpointLambdas(this, 'PodcastHandler', {
        workshopLanguage: props.workshopLanguage,
        podcastBucket: podcastBucket,
        podcastEpisodeTable: podcastEpisodeTable,
        transcribeStateMachine: transcribeStateMachine,
      }),
    });
    new cdk.CfnOutput(this, 'APIUrl', {
      value: frontend.httpApi.apiEndpoint,
    });
  }
}

function makeTranscribeAccessBucketRole(
  scope: cdk.Construct,
  id: string,
  bucket: s3.IBucket
): iam.IRole {
  return new iam.Role(scope, id, {
    assumedBy: new iam.ServicePrincipal('transcribe.amazonaws.com'),
    inlinePolicies: {
      ArchiveStreamRolePolicy: iamUtils.makePolicyDocument({
        statements: [
          iamUtils.makePolicyStatement({
            effect: iam.Effect.ALLOW,
            actions: [
              's3:AbortMultipartUpload',
              's3:GetBucketLocation',
              's3:GetObject',
              's3:ListBucket',
              's3:ListBucketMultipartUploads',
              's3:PutObject',
            ],
            resources: [bucket.bucketArn, bucket.bucketArn + '/*'],
          }),
        ],
      }),
    },
  });
}

interface podcastHandlers {
  listPodcastsFn: lambda.IFunction;
  addPodcastFn: lambda.IFunction;
  getPodcastFn: lambda.IFunction;
  playPodcastFn: lambda.IFunction;
}

interface makeApiEndpointLambdasProps {
  podcastBucket: s3.IBucket;
  podcastEpisodeTable: ddb.ITable;
  transcribeStateMachine: sfn.IStateMachine;

  workshopLanguage: WorkshopLanguage;
}

const commonStaticLambdaEnvs = {
  [ENV_KEY_PODCAST_DATA_KEY_PREFIX]: PODCAST_DATA_KEY_PREFIX,
  [ENV_KEY_MAX_NUM_EPISODE_IMPORT]: MAX_NUM_EPISODE_IMPORT,
  AWS_RETRY_MODE: 'standard',
  AWS_MAX_ATTEMPTS: '3',
};

function makeApiEndpointLambdas(
  scope: cdk.Construct,
  id: string,
  props: makeApiEndpointLambdasProps
): podcastHandlers {
  const listPodcastsId = id + 'ListPodcasts';
  const getPodcastId = id + 'GetPodcast';
  const playPodcastId = id + 'PlayPodcast';

  const commonProps = {
    environment: {
      [ENV_KEY_TRANSCRIBE_STATEMACHINE_ARN]:
        props.transcribeStateMachine.stateMachineArn,
      [ENV_KEY_PODCAST_EPISODE_TABLE_NAME]: props.podcastEpisodeTable.tableName,
      [ENV_KEY_PODCAST_DATA_BUCKET_NAME]: props.podcastBucket.bucketName,
      ...commonStaticLambdaEnvs,
    },
    memorySize: 1024,
    timeout: cdk.Duration.seconds(60),
  };

  // Lambda handler implementations common across all languages
  const addPodcastFn = new lambda.Function(scope, id + 'AddPodcast', {
    runtime: lambda.Runtime.GO_1_X,
    handler: 'main',
    code: lambda.Code.fromAsset('lambda/go/add-podcasts'),
    ...commonProps,
  });

  let handlers: podcastHandlers;
  switch (props.workshopLanguage) {
    case WorkshopLanguage.Go:
      handlers = {
        // Common handlers
        addPodcastFn: addPodcastFn,

        // language specific handlers
        listPodcastsFn: new lambda.Function(scope, listPodcastsId, {
          runtime: lambda.Runtime.GO_1_X,
          handler: 'main',
          code: lambda.Code.fromAsset('lambda/go/list-podcasts'),
          ...commonProps,
        }),
        getPodcastFn: new lambda.Function(scope, getPodcastId, {
          runtime: lambda.Runtime.GO_1_X,
          handler: 'main',
          code: lambda.Code.fromAsset('lambda/go/get-podcast'),
          ...commonProps,
        }),
        playPodcastFn: new lambda.Function(scope, playPodcastId, {
          runtime: lambda.Runtime.GO_1_X,
          handler: 'main',
          code: lambda.Code.fromAsset('lambda/go/play-podcast'),
          ...commonProps,
        }),
      };
      break;

    case WorkshopLanguage.Python:
      handlers = {
        // Common handlers
        addPodcastFn: addPodcastFn,

        // language specific handlers
        listPodcastsFn: new lambda.Function(scope, listPodcastsId, {
          runtime: lambda.Runtime.PYTHON_3_8,
          handler: 'list_podcasts.lambda_handler',
          code: lambda.Code.fromAsset('lambda/python'),
          ...commonProps,
        }),
        getPodcastFn: new lambda.Function(scope, getPodcastId, {
          runtime: lambda.Runtime.PYTHON_3_8,
          handler: 'get_podcast.lambda_handler',
          code: lambda.Code.fromAsset('lambda/python'),
          ...commonProps,
        }),
        playPodcastFn: new lambda.Function(scope, playPodcastId, {
          runtime: lambda.Runtime.PYTHON_3_8,
          handler: 'play_podcast.lambda_handler',
          code: lambda.Code.fromAsset('lambda/python'),
          ...commonProps,
        }),
      };
      break;

    case WorkshopLanguage.Java:
      handlers = {
        // Common handlers
        addPodcastFn: addPodcastFn,

        // language specific handlers
        listPodcastsFn: new lambda.Function(scope, listPodcastsId, {
          runtime: lambda.Runtime.JAVA_11,
          handler: 'com.amazonaws.workshop.ListPodcasts::handleRequest',
          code: lambda.Code.fromAsset('lambda/java'),
          ...commonProps,
        }),
        getPodcastFn: new lambda.Function(scope, getPodcastId, {
          runtime: lambda.Runtime.JAVA_11,
          handler: 'com.amazonaws.workshop.GetPodcast::handleRequest',
          code: lambda.Code.fromAsset('lambda/java'),
          ...commonProps,
        }),
        playPodcastFn: new lambda.Function(scope, playPodcastId, {
          runtime: lambda.Runtime.JAVA_11,
          handler: 'com.amazonaws.workshop.PlayPodcast::handleRequest',
          code: lambda.Code.fromAsset('lambda/java'),
          ...commonProps,
        }),
      };
      break;

    case WorkshopLanguage.Javascript:
      handlers = {
        // Common handlers
        addPodcastFn: addPodcastFn,

        // language specific handlers
        listPodcastsFn: new lambda_nodejs.NodejsFunction(scope, listPodcastsId, {
          entry: 'lambda/javascript/list-podcasts.js',
          bundling: getNodejsBundlingOptions(listPodcastsId),
          ...commonProps,
        }),
        getPodcastFn: new lambda_nodejs.NodejsFunction(scope, getPodcastId, {
          entry: 'lambda/javascript/get-podcast.js',
          bundling: getNodejsBundlingOptions(getPodcastId),
          ...commonProps,
        }),
        playPodcastFn: new lambda_nodejs.NodejsFunction(scope, playPodcastId, {
          entry: 'lambda/javascript/play-podcast.js',
          bundling: getNodejsBundlingOptions(playPodcastId),
          ...commonProps,
        }),
      };
      break;

    default:
      throw new Error(`unknown workshop language, ${props.workshopLanguage}`);
  }

  //------------------------------
  // List Podcast
  //------------------------------
  handlers.listPodcastsFn.addToRolePolicy(
    iamUtils.makePolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: ['dynamodb:Scan'],
      resources: [props.podcastEpisodeTable.tableArn],
    })
  );

  //------------------------------
  // Add Podcast
  //------------------------------
  handlers.addPodcastFn.addToRolePolicy(
    iamUtils.makePolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: ['states:StartExecution'],
      resources: [props.transcribeStateMachine.stateMachineArn],
    })
  );
  handlers.addPodcastFn.addToRolePolicy(
    iamUtils.makePolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: [
        'dynamodb:BatchGetItem',
        'dynamodb:BatchWriteItem',
        'dynamodb:UpdateItem',
      ],
      resources: [props.podcastEpisodeTable.tableArn],
    })
  );

  //------------------------------
  // Get Podcasts
  //------------------------------
  if (handlers.getPodcastFn.role) {
    props.podcastEpisodeTable.grantReadData(handlers.getPodcastFn.role);
  }

  //------------------------------
  // Play Podcast
  //------------------------------
  handlers.playPodcastFn.addToRolePolicy(
    iamUtils.makePolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: ['dynamodb:GetItem'],
      resources: [props.podcastEpisodeTable.tableArn],
    })
  );
  handlers.playPodcastFn.addToRolePolicy(
    iamUtils.makePolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: [
        's3:AbortMultipartUpload',
        's3:GetBucketLocation',
        's3:GetObject',
        's3:ListBucket',
        's3:ListBucketMultipartUploads',
        's3:PutObject',
      ],
      resources: [
        props.podcastBucket.bucketArn,
        props.podcastBucket.bucketArn + '/*',
      ],
    })
  );

  return handlers;
}

interface transcribeStatemachineHandlers {
  updateEpisodeStatus: lambda.IFunction;
  uploadPodcast: lambda.IFunction;
  startTranscription: lambda.IFunction;
  checkTranscription: lambda.IFunction;
  processTranscription: lambda.IFunction;
}

interface makeTranscribeStatemachineLambdasProps {
  podcastBucket: s3.IBucket;
  podcastEpisodeTable: ddb.ITable;
  transcribeAccessRole: iam.IRole;

  workshopLanguage: WorkshopLanguage;
}

function makeTranscribeStatemachineLambdas(
  scope: cdk.Construct,
  id: string,
  props: makeTranscribeStatemachineLambdasProps
): transcribeStatemachineHandlers {
  const uploadPodcastId = id + 'UploadPodcast';
  const startTranscriptionId = id + 'StartTranscription';
  const checkTranscriptionId = id + 'CheckTranscription';
  const processTranscriptionId = id + 'CleanupTranscription';

  const commonProps = {
    environment: {
      [ENV_KEY_PODCAST_EPISODE_TABLE_NAME]: props.podcastEpisodeTable.tableName,
      [ENV_KEY_PODCAST_DATA_BUCKET_NAME]: props.podcastBucket.bucketName,
      [ENV_KEY_TRANSCRIBE_ACCESS_ROLE_ARN]: props.transcribeAccessRole.roleArn,
      ...commonStaticLambdaEnvs,
    },
    memorySize: 1024,
    timeout: cdk.Duration.seconds(60),
  };

  // Lambda handler implementations common across all languages
  const updateEpisodeStatus = new lambda.Function(
    scope,
    id + 'UpdateEpisodeStatus',
    {
      runtime: lambda.Runtime.GO_1_X,
      handler: 'main',
      code: lambda.Code.fromAsset('lambda/go/update-episode-status'),
      ...commonProps,
    }
  );

  const startTranscription = new lambda.Function(scope, startTranscriptionId, {
    runtime: lambda.Runtime.GO_1_X,
    handler: 'main',
    code: lambda.Code.fromAsset('lambda/go/start-transcription'),
    ...commonProps,
  });

  const checkTranscription = new lambda.Function(scope, checkTranscriptionId, {
    runtime: lambda.Runtime.GO_1_X,
    handler: 'main',
    code: lambda.Code.fromAsset('lambda/go/check-transcription'),
    ...commonProps,
  });

  const processTranscription = new lambda.Function(
    scope,
    processTranscriptionId,
    {
      runtime: lambda.Runtime.GO_1_X,
      handler: 'main',
      code: lambda.Code.fromAsset('lambda/go/process-transcription'),
      ...commonProps,
    }
  );

  let handlers: transcribeStatemachineHandlers;
  switch (props.workshopLanguage) {
    case WorkshopLanguage.Go:
      handlers = {
        // Common handlers
        updateEpisodeStatus: updateEpisodeStatus,
        startTranscription: startTranscription,
        checkTranscription: checkTranscription,
        processTranscription: processTranscription,

        // language specific handlers
        uploadPodcast: new lambda.Function(scope, uploadPodcastId, {
          runtime: lambda.Runtime.GO_1_X,
          handler: 'main',
          code: lambda.Code.fromAsset('lambda/go/upload-podcast'),
          ...commonProps,
        }),
      };
      break;

    case WorkshopLanguage.Python:
      handlers = {
        // Common handlers
        updateEpisodeStatus: updateEpisodeStatus,
        startTranscription: startTranscription,
        checkTranscription: checkTranscription,
        processTranscription: processTranscription,

        // language specific handlers
        uploadPodcast: new lambda.Function(scope, uploadPodcastId, {
          runtime: lambda.Runtime.PYTHON_3_8,
          handler: 'upload_podcast.lambda_handler',
          code: lambda.Code.fromAsset('lambda/python'),
          ...commonProps,
        }),
      };
      break;

    case WorkshopLanguage.Java:
      handlers = {
        // Common handlers
        updateEpisodeStatus: updateEpisodeStatus,
        startTranscription: startTranscription,
        checkTranscription: checkTranscription,
        processTranscription: processTranscription,

        // language specific handlers
        uploadPodcast: new lambda.Function(scope, uploadPodcastId, {
          runtime: lambda.Runtime.JAVA_11,
          handler: 'com.amazonaws.workshop.UploadPodcast::handleRequest',
          code: lambda.Code.fromAsset('lambda/java'),
          ...commonProps,
        }),
      };
      break;

    case WorkshopLanguage.Javascript:
      handlers = {
        // Common handlers
        updateEpisodeStatus: updateEpisodeStatus,
        startTranscription: startTranscription,
        checkTranscription: checkTranscription,
        processTranscription: processTranscription,

        // language specific handlers
        uploadPodcast: new lambda_nodejs.NodejsFunction(scope, uploadPodcastId, {
          entry: 'lambda/javascript/upload-podcast.js',
          bundling: getNodejsBundlingOptions(uploadPodcastId),
          ...commonProps,
        }),
      };
      break;

    default:
      throw new Error(`unknown workshop language, ${props.workshopLanguage}`);
  }

  //------------------------------
  // Update Episode Status
  //------------------------------
  handlers.updateEpisodeStatus.addToRolePolicy(
    iamUtils.makePolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: ['dynamodb:UpdateItem'],
      resources: [props.podcastEpisodeTable.tableArn],
    })
  );

  //------------------------------
  // Upload Podcast
  //------------------------------
  handlers.uploadPodcast.addToRolePolicy(
    iamUtils.makePolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: [
        's3:AbortMultipartUpload',
        's3:GetBucketLocation',
        's3:GetObject',
        's3:ListBucket',
        's3:ListBucketMultipartUploads',
        's3:PutObject',
      ],
      resources: [
        props.podcastBucket.bucketArn,
        props.podcastBucket.bucketArn + '/*',
      ],
    })
  );

  //------------------------------
  // Start Transcription
  //------------------------------
  handlers.startTranscription.addToRolePolicy(
    iamUtils.makePolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: ['transcribe:StartTranscriptionJob'],
      resources: ['*'],
    })
  );
  handlers.startTranscription.addToRolePolicy(
    iamUtils.makePolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: ['iam:PassRole'],
      resources: ['*'],
    })
  );

  //------------------------------
  // Check Transcription
  //------------------------------
  handlers.checkTranscription.addToRolePolicy(
    iamUtils.makePolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: ['transcribe:GetTranscriptionJob'],
      resources: ['*'],
    })
  );

  //------------------------------
  // Process Transcription
  //------------------------------
  handlers.processTranscription.addToRolePolicy(
    iamUtils.makePolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: ['transcribe:GetTranscriptionJob'],
      resources: ['*'],
    })
  );
  handlers.processTranscription.addToRolePolicy(
    iamUtils.makePolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: [
        's3:AbortMultipartUpload',
        's3:GetBucketLocation',
        's3:GetObject',
        's3:ListBucket',
        's3:ListBucketMultipartUploads',
        's3:PutObject',
      ],
      resources: [
        props.podcastBucket.bucketArn,
        props.podcastBucket.bucketArn + '/*',
      ],
    })
  );
  handlers.processTranscription.addToRolePolicy(
    iamUtils.makePolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: ['dynamodb:PutItem'],
      resources: [props.podcastEpisodeTable.tableArn],
    })
  );

  return handlers;
}
