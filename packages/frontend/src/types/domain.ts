// 命令状态机
export enum AssetCommandStatus {
  QUEUED = 'queued',
  SENT = 'sent',
  ACKED = 'acked',
  FAILED = 'failed',
  TIMEOUT = 'timeout',
  CANCELLED = 'cancelled'
}
