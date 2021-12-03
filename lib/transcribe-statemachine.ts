import * as cdk from 'monocdk';
import * as lambda from 'monocdk/aws-lambda';
import * as logs from 'monocdk/aws-logs';
import * as sfn from 'monocdk/aws-stepfunctions';
import * as sfnTasks from 'monocdk/aws-stepfunctions-tasks';

export interface TranscribeStateMachineProps {
  updateEpisodeStatus: lambda.IFunction;

  uploadPodcast: lambda.IFunction;
  startTranscription: lambda.IFunction;
  checkTranscription: lambda.IFunction;
  processTranscription: lambda.IFunction;
}

/**
 * Provides CDK Construct for a Step Function state machine for downloading an
 * podcast from an RSS field, uploading the audio to S3 bucket, and kicking off
 * an job with Amazon Transcribe to transcribe the audio to text.
 */
export class TranscribeStateMachine extends cdk.Construct {
  public readonly stateMachine: sfn.IStateMachine;

  constructor(
    scope: cdk.Construct,
    id: string,
    props: TranscribeStateMachineProps
  ) {
    super(scope, id);

    const failureStep = makeUpdateStatusState(
      this,
      'Failure',
      props.updateEpisodeStatus
    ).next(new sfn.Fail(this, 'Failure'));

    const uploadLambdaStep = new sfnTasks.LambdaInvoke(
      this,
      'UploadPodcastStep',
      {
        lambdaFunction: props.uploadPodcast,
        payloadResponseOnly: true,
        outputPath: '$.episode',
        resultPath: '$.episode',
      }
    ).addCatch(failureStep, {
      errors: ['States.TaskFailed'],
      resultPath: '$.taskFailed',
    });

    const startTranscriptionStep = new sfnTasks.LambdaInvoke(
      this,
      'StartTranscriptionStep',
      {
        lambdaFunction: props.startTranscription,
        payloadResponseOnly: true,
        outputPath: '$.episode',
        resultPath: '$.episode',
      }
    ).addCatch(failureStep, {
      errors: ['States.TaskFailed'],
      resultPath: '$.taskFailed',
    });

    const checkTranscriptionStep = new sfnTasks.LambdaInvoke(
      this,
      'CheckTranscriptionStep',
      {
        lambdaFunction: props.checkTranscription,
        payloadResponseOnly: true,
        resultPath: '$.transcribeStatus',
      }
    ).addCatch(failureStep, {
      errors: ['States.TaskFailed'],
      resultPath: '$.taskFailed',
    });

    const processTranscriptionStep = new sfnTasks.LambdaInvoke(
      this,
      'ProcessTranscriptionStep',
      {
        lambdaFunction: props.processTranscription,
        payloadResponseOnly: true,
        //resultPath: '$.episode',
        resultPath: sfn.JsonPath.DISCARD,
      }
    ).addCatch(failureStep, {
      errors: ['States.TaskFailed'],
      resultPath: '$.taskFailed',
    });

    const processingStep = makeUpdateStatusState(
      this,
      'Processing',
      props.updateEpisodeStatus
    )
      .next(processTranscriptionStep)
      .next(makeUpdateStatusState(this, 'Complete', props.updateEpisodeStatus))
      .next(new sfn.Succeed(this, 'Complete'));

    const definition = makeUpdateStatusState(
      this,
      'Uploading',
      props.updateEpisodeStatus
    )
      .next(uploadLambdaStep)
      .next(
        makeUpdateStatusState(this, 'Transcribing', props.updateEpisodeStatus)
      )
      .next(startTranscriptionStep)
      .next(checkTranscriptionStep)
      .next(
        new ChoiceTask(this, 'IsTranscribeComplete', {
          when: [
            {
              condition: sfn.Condition.and(
                sfn.Condition.isPresent('$.transcribeStatus.status'),
                sfn.Condition.stringEquals(
                  '$.transcribeStatus.status',
                  'COMPLETED'
                )
              ),
              next: processingStep,
            },
            {
              condition: sfn.Condition.and(
                sfn.Condition.isPresent('$.transcribeStatus.status'),
                sfn.Condition.stringEquals(
                  '$.transcribeStatus.status',
                  'FAILED'
                )
              ),
              next: failureStep,
            },
          ],
          otherwise: new sfn.Wait(this, 'WaitForTranscription', {
            time: sfn.WaitTime.duration(cdk.Duration.seconds(30)),
          }).next(checkTranscriptionStep),
        }).afterwards()
      );

    this.stateMachine = new sfn.StateMachine(this, 'StateMachine', {
      definition: definition,
    });
  }
}

interface ChoiceConditionProps {
  readonly condition: sfn.Condition;
  readonly next: sfn.IChainable;
}

interface ChoiceTaskProps extends sfn.ChoiceProps {
  readonly when: ChoiceConditionProps[];
  readonly otherwise?: sfn.IChainable;
}

export class ChoiceTask {
  choice: sfn.Choice;

  constructor(scope: cdk.Construct, id: string, props: ChoiceTaskProps) {
    this.choice = new sfn.Choice(scope, id, props);

    props.when.forEach((w) => {
      this.choice = this.choice.when(w.condition, w.next);
    });

    const otherwise: sfn.IChainable =
      props.otherwise || new sfn.Pass(scope, id + 'Pass');
    this.choice = this.choice.otherwise(otherwise);
  }

  // Return a Chain that contains all reachable end states from this Choice.
  // Use this to combine all possible choice paths back.
  public afterwards(options?: sfn.AfterwardsOptions): sfn.IChainable {
    return this.choice.afterwards(options);
  }
}

function makeUpdateStatusState(
  scope: cdk.Construct,
  status: string,
  fn: lambda.IFunction
): sfn.TaskStateBase {
  const id = 'UpdateStatus' + status + 'Lambda';

  return new sfnTasks.LambdaInvoke(scope, id, {
    lambdaFunction: fn,
    payload: sfn.TaskInput.fromObject({
      id: sfn.JsonPath.stringAt('$.episode.id'),
      status: status.toLowerCase(),
    }),
    payloadResponseOnly: true,
    resultPath: '$.episode.status',
    //resultPath: sfn.JsonPath.DISCARD,
  });
}
