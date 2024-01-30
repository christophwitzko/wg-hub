"use client";
import { useAuth } from "@/lib/auth";
import { usePeers } from "@/lib/api";

export default function Peers() {
  const auth = useAuth();
  const { data, error, isLoading } = usePeers(auth.token);
  return (
    <div>
      <h1>Peers</h1>
      {error && <div>Error: {error.message}</div>}
      {isLoading && <div>Loading...</div>}
      {data && (
        <ul>
          {data?.map((peer) => <li key={peer.publicKey}>{peer.publicKey}</li>)}
        </ul>
      )}
    </div>
  );
}
