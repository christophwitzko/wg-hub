"use client";

import {
  AlertCircle,
  ClipboardCopy,
  FileCog,
  Plus,
  QrCode,
  Shuffle,
} from "lucide-react";
import {
  useCallback,
  useEffect,
  useState,
  MouseEvent,
  useMemo,
  useRef,
} from "react";

import { useAuth } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { GeneratedPeer, generatePeer, Hub, useHub } from "@/lib/api";
import { Input } from "@/components/ui/input";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Code } from "@/components/code";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";

const generatePeerFormSchema = z.object({
  allowedIP: z.string().ip({ version: "v4" }),
});

function generateConfig(hub?: Hub | null, peer?: GeneratedPeer | null) {
  if (!hub || !peer) return "";
  return `[Interface]
Address = ${peer.allowedIP}
PrivateKey = ${peer.privateKey}

[Peer]
PublicKey = ${hub.publicKey}
AllowedIPs = ${hub.hubNetwork}
Endpoint = 127.0.0.1:${hub.port} # TODO
PersistentKeepalive = 25
`;
}

export function GeneratePeer() {
  const [open, setOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [generatedPeer, setGeneratedPeer] = useState<GeneratedPeer | null>(
    null,
  );
  const [showQR, setShowQR] = useState(false);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const auth = useAuth();
  const hub = useHub(auth.token);

  const form = useForm<z.infer<typeof generatePeerFormSchema>>({
    resolver: zodResolver(generatePeerFormSchema),
    defaultValues: {
      allowedIP: "",
    },
  });
  function onSubmit(values: z.infer<typeof generatePeerFormSchema>) {
    setError("");
    setIsLoading(true);
    generatePeer(auth.token, values)
      .then((newPeer) => {
        setGeneratedPeer(newPeer);
        toast("Peer generated");
      })
      .catch((error) => {
        setError(error.message);
      })
      .finally(() => {
        setIsLoading(false);
      });
  }

  useEffect(() => {
    if (open) {
      form.reset();
      setGeneratedPeer(null);
      setShowQR(false);
      setError("");
    }
  }, [form, open]);

  const randomIP = useCallback(
    (e: MouseEvent<HTMLButtonElement>) => {
      e.preventDefault();
      hub
        .mutate()
        .then(() =>
          form.setValue(
            "allowedIP",
            (hub.data?.randomFreeIP || "").slice(0, -3),
          ),
        );
    },
    [form, hub],
  );

  const peerConfig = useMemo(
    () => generateConfig(hub.data, generatedPeer),
    [hub.data, generatedPeer],
  );

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline">
          Generate New Peer <Plus className="ml-2 size-4" />
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-fit">
        <DialogHeader>
          <DialogTitle>Generate New Peer</DialogTitle>
          <DialogDescription>
            Generate and add a new peer to wg-hub.
          </DialogDescription>
        </DialogHeader>
        {generatedPeer ? (
          <div className="flex gap-2 flex-col">
            {showQR ? (
              <canvas ref={canvasRef} />
            ) : (
              <Code lang="ini" value={peerConfig}></Code>
            )}
            <DialogFooter>
              <Button
                variant="secondary"
                onClick={() => setShowQR((prev) => !prev)}
              >
                {showQR ? (
                  <FileCog className="mr-2 size-4" />
                ) : (
                  <QrCode className="mr-2 size-4" />
                )}
                {showQR ? "Show Config" : "Show QR Code"}
              </Button>
              {!showQR && (
                <Button
                  onClick={() => {
                    navigator.clipboard.writeText(peerConfig);
                    toast("Copied to clipboard");
                  }}
                >
                  <ClipboardCopy className="mr-2 size-4" />
                  Copy to Clipboard
                </Button>
              )}
            </DialogFooter>
          </div>
        ) : (
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)}>
              <div className="grid gap-4 py-4">
                <FormField
                  control={form.control}
                  name="allowedIP"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Allowed IP</FormLabel>
                      <div className="flex gap-2">
                        <FormControl>
                          <Input {...field} data-1p-ignore />
                        </FormControl>
                        <Button
                          size="icon"
                          variant="outline"
                          className="flex-shrink-0"
                          onClick={randomIP}
                        >
                          <Shuffle className="size-5" />
                        </Button>
                      </div>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                {error && (
                  <Alert variant="destructive">
                    <AlertCircle className="size-4" />
                    <AlertTitle>Error</AlertTitle>
                    <AlertDescription>{error}</AlertDescription>
                  </Alert>
                )}
              </div>
              <DialogFooter>
                <Button type="submit" disabled={isLoading}>
                  Generate
                </Button>
              </DialogFooter>
            </form>
          </Form>
        )}
      </DialogContent>
    </Dialog>
  );
}
