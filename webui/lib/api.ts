"use client";

import useSWR from "swr";

async function fetchAPI(
  method: string,
  path: string,
  token: string = "",
  body: any = undefined,
) {
  const res = await fetch(`/api/${path}`, {
    method,
    headers: {
      Authorization: token ? `Bearer ${token}` : "",
      "Content-Type": "application/json",
    },
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) {
    // handle client errors
    if (res.status >= 400 && res.status < 500) {
      const err = await res.json();
      throw new Error(err.error);
    }
    // handle server errors
    throw new Error(res.statusText);
  }
  return await res.json();
}

export async function createToken(
  username: string,
  password: string,
): Promise<string> {
  const authRes = await fetchAPI("POST", "auth", "", { username, password });
  if (!authRes.token) {
    throw new Error("Invalid response from server.");
  }
  return authRes.token;
}

export type User = {
  username: string;
  iat: number;
  exp: number;
};
export async function getUser(token: string): Promise<User> {
  return await fetchAPI("GET", "auth", token);
}

export type Peer = {
  publicKey: string;
  allowedIP: string;
  endpoint: string;
  lastHandshake: number;
  txBytes: number;
  rxBytes: number;
  isHub: boolean;
  isRequester: boolean;
};
export function usePeers(token: string) {
  return useSWR<Peer[]>("peers", () => fetchAPI("GET", "peers", token), {
    refreshInterval: 1000,
    fallbackData: [],
  });
}

export type Config = {
  config: string;
};

export function useConfig(token: string) {
  return useSWR<Config>("config", () => fetchAPI("GET", "config", token), {
    refreshInterval: 1000,
  });
}

export type AddedPeer = {
  allowedIP: string;
  hubNetwork: string;
};
export async function addPeer(
  token: string,
  peer: { publicKey: string; allowedIP: string },
): Promise<AddedPeer> {
  return await fetchAPI("PUT", `peers/${peer.publicKey}`, token, {
    allowedIP: peer.allowedIP,
  });
}

export async function removePeer(
  token: string,
  publicKey: string,
): Promise<void> {
  await fetchAPI("DELETE", `peers/${publicKey}`, token);
}

export type Hub = {
  publicKey: string;
  port: number;
  hubNetwork: string;
  randomFreeIP: string;
  externalIP: string;
};

export function useHub(token: string) {
  return useSWR<Hub>("hub", () => fetchAPI("GET", "hub", token), {
    refreshInterval: 5000,
  });
}

export type GeneratedPeer = {
  privateKey: string;
  publicKey: string;
  allowedIP: string;
  hubNetwork: string;
};
export async function generatePeer(
  token: string,
  peer: { allowedIP: string },
): Promise<GeneratedPeer> {
  return await fetchAPI("POST", `peers`, token, {
    allowedIP: peer.allowedIP,
  });
}
