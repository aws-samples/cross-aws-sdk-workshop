import * as cdk from 'monocdk';
import * as apiv2 from 'monocdk/aws-apigatewayv2';
import * as apiv2Integ from 'monocdk/aws-apigatewayv2-integrations';
import * as lambda from 'monocdk/aws-lambda';

export interface ApiGatewayFrontEndProps {
  listPodcastsFn: lambda.IFunction;
  addPodcastFn: lambda.IFunction;
  getPodcastFn: lambda.IFunction;
  playPodcastFn: lambda.IFunction;
}

export class ApiGatewayFrontend extends cdk.Construct {
  public readonly httpApi: apiv2.HttpApi;

  constructor(
    scope: cdk.Construct,
    id: string,
    props: ApiGatewayFrontEndProps
  ) {
    super(scope, id);

    this.httpApi = new apiv2.HttpApi(this, 'ApiGateWayHttpApi');

    this.httpApi.addRoutes({
      path: '/podcast',
      methods: [apiv2.HttpMethod.GET],
      integration: new apiv2Integ.LambdaProxyIntegration({
        handler: props.listPodcastsFn,
      }),
    });

    this.httpApi.addRoutes({
      path: '/podcast',
      methods: [apiv2.HttpMethod.POST],
      integration: new apiv2Integ.LambdaProxyIntegration({
        handler: props.addPodcastFn,
      }),
    });

    this.httpApi.addRoutes({
      path: '/podcast/{id}',
      methods: [apiv2.HttpMethod.GET],
      integration: new apiv2Integ.LambdaProxyIntegration({
        handler: props.getPodcastFn,
      }),
    });

    this.httpApi.addRoutes({
      path: '/podcast/{id}/play',
      methods: [apiv2.HttpMethod.GET],
      integration: new apiv2Integ.LambdaProxyIntegration({
        handler: props.playPodcastFn,
      }),
    });
  }
}
