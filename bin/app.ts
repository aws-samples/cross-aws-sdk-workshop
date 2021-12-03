#!/usr/bin/env node
import fs from 'fs';
import path from 'path';
import 'source-map-support/register';
import * as cdk from 'monocdk';
import * as ws from '../lib/cdk-stack';

interface WorkshopConfig {
  account: string,
  region: string,
  language: string,
}

function parseWorkshopConfigFile(filename: string): WorkshopConfig {
  try {
    fs.accessSync(filename, fs.constants.R_OK);
  } catch(err) {
    throw new Error(`workshop configuration file, ${filename} not found. Run 'make bootstrap-config'`); 
  }

  const fileData: unknown = JSON.parse(fs.readFileSync(filename, 'utf-8'));
  if (!isWorkshopConfig(fileData)) {
    throw new Error(`unknown workshop configuration file data, ${filename}`);
  }

  return fileData;
}
export function isWorkshopConfig(o: unknown): o is WorkshopConfig {
  return (o as WorkshopConfig).account !== undefined;
}

const config = parseWorkshopConfigFile(path.resolve(__dirname, '..', '.workshop-config.json'));

const app = new cdk.App();
new ws.CdkStack(app, 'AwsCrossSdkWorkshop', {
  env: { 
    account: config.account,
    region: config.region
  },
  workshopLanguage: ws.parseWorkshopLanguage(config.language),

  /* For more information, see https://docs.aws.amazon.com/cdk/latest/guide/environments.html */
});
