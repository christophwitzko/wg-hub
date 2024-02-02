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
};
export function usePeers(token: string) {
  return useSWR<Peer[]>("peers", () => fetchAPI("GET", "peers", token), {
    refreshInterval: 1000,
  });
}