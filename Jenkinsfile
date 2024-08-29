containerReleasePipeline(
  k8sAgentParams: [
    name: 'go',
    arch: 'amd64',
    defaultContainer: 'go',
    goVersion: '1.14',
  ],
  ecrPushParams: [
    staging: [
      account: '888026007073',
      regions: 'us-east-1',
    ],
    prod: [
      account: '716756199562',
      regions: 'us-east-1, eu-central-1, ap-northeast-1, ap-southeast-2, ca-central-1, us-east-2, us-west-2',
    ]
  ]
)
