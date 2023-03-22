export interface Device {
  index: number;
  name: string;
  hardware_addr: string;
  addrs: Addr[];
}

export interface Addr {
  network_name: string;
  ip: string;
  prefix: number;
  ip_type_str: string;
}

export function formatAddr(addr: Addr): string {
  return `${addr.network_name} ${addr.ip}/${addr.prefix}  ${addr.ip_type_str}`;
}

export interface Packet {
  device_index: number;
  timestamp: Date;
  length: number;
  capture_length: number;
  layers: Layer[];
  src: string;
  dst: string;
  packet_type: string;
}

export interface Layer {
  layer_type: string;
  src: string;
  dst: string;
  payload: string;
  dump: string;
  fields: Record<string, string>;
}

export interface PcapConfig {
  device: string;
  filter: string;

  // TODO: implement these
  // snaplen: number;
  // promisc: boolean;
  // timeout: number;
}
