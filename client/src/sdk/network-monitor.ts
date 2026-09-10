export type NetworkQuality = 'unknown' | 'good' | 'fair' | 'poor'

export interface NetworkQualitySample {
  quality: NetworkQuality
  rtt: number | null
  packetLoss: number
  packetsReceived: number
  packetsLost: number
}

export interface NetworkStatsPeer {
  getStats: () => Promise<RTCStatsReport>
}

export interface NetworkQualityMonitorOptions {
  intervalMs?: number
  onSample: (sample: NetworkQualitySample) => void
}

/**
 * Samples inbound video and candidate-pair statistics without depending on
 * Vuex or UI components. A missing peer or transient getStats failure is
 * intentionally ignored; the next sample can recover the display.
 */
export class NetworkQualityMonitor {
  private readonly intervalMs: number
  private readonly onSample: (sample: NetworkQualitySample) => void
  private peer?: NetworkStatsPeer
  private timer?: number
  private previous?: { packetsReceived: number; packetsLost: number }

  constructor(options: NetworkQualityMonitorOptions) {
    this.intervalMs = options.intervalMs ?? 5000
    this.onSample = options.onSample
  }

  start(peer: NetworkStatsPeer) {
    this.stop()
    this.peer = peer
    this.previous = undefined
    void this.sample()
    this.timer = window.setInterval(() => void this.sample(), this.intervalMs)
  }

  stop() {
    if (this.timer !== undefined) {
      window.clearInterval(this.timer)
      this.timer = undefined
    }
    this.peer = undefined
    this.previous = undefined
  }

  private async sample() {
    if (!this.peer) {
      return
    }

    try {
      const stats = await this.peer.getStats()
      let packetsReceived = 0
      let packetsLost = 0
      let rtt: number | null = null

      stats.forEach((stat: any) => {
        if (stat.type === 'inbound-rtp' && (stat.kind === 'video' || stat.mediaType === 'video')) {
          packetsReceived += Number(stat.packetsReceived || 0)
          packetsLost += Number(stat.packetsLost || 0)
        }

        if (
          stat.type === 'candidate-pair' &&
          (stat.state === 'succeeded' || stat.nominated === true) &&
          typeof stat.currentRoundTripTime === 'number'
        ) {
          rtt = stat.currentRoundTripTime * 1000
        }
      })

      const previous = this.previous
      this.previous = { packetsReceived, packetsLost }
      const receivedDelta = previous ? Math.max(0, packetsReceived - previous.packetsReceived) : packetsReceived
      const lostDelta = previous ? Math.max(0, packetsLost - previous.packetsLost) : packetsLost
      const totalPackets = receivedDelta + lostDelta
      const packetLoss = totalPackets > 0 ? lostDelta / totalPackets : 0
      const quality = classifyNetworkQuality(rtt, packetLoss, totalPackets > 0 || rtt !== null)

      this.onSample({
        quality,
        rtt: rtt === null ? null : Math.round(rtt),
        packetLoss,
        packetsReceived,
        packetsLost,
      })
    } catch {
      // getStats is best effort; a temporary failure must not affect the media session.
    }
  }
}

export function classifyNetworkQuality(rtt: number | null, packetLoss: number, hasStats: boolean): NetworkQuality {
  if (!hasStats) {
    return 'unknown'
  }

  if ((rtt !== null && rtt > 350) || packetLoss > 0.08) {
    return 'poor'
  }

  if ((rtt !== null && rtt > 180) || packetLoss > 0.03) {
    return 'fair'
  }

  return 'good'
}
