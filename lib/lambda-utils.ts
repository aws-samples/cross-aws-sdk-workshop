import { BundlingOptions } from 'monocdk/aws-lambda-nodejs';

/**
 * SAM CLI(sam-beta-cdk) deployment requires package.json to exist in the 
 * deployment package. However bundled Node.js function doesn't have a 
 * package.json. This workaround generates an empty package.json for these
 * functions.
 */
export const getNodejsBundlingOptions = (id: string): BundlingOptions => ({
  commandHooks: {
    beforeInstall: () => [],
    afterBundling: (_, outputDir) => [
      `echo '${JSON.stringify({
        name: id,
        version: '0.1.0',
        private: true,
        dependencies: {},
      }, undefined, 2)}' > ${outputDir}/package.json`,
    ],
    beforeBundling: () => [],
  }
});
