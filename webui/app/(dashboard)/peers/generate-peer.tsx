"use client";

import { AlertCircle, Plus, Shuffle } from "lucide-react";
import { useCallback, useEffect, useState, MouseEvent } from "react";

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
import { GeneratedPeer, generatePeer, useHub } from "@/lib/api";
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
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";

const generatePeerFormSchema = z.object({
  allowedIP: z.string().ip({ version: "v4" }),
});

export function GeneratePeer() {
  const [open, setOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [generatedPeer, setGeneratedPeer] = useState<GeneratedPeer | null>(
    null,
  );
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

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline">
          Generate New Peer <Plus className="ml-2 size-4" />
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Generate New Peer</DialogTitle>
          <DialogDescription>
            Generate and add a new peer to wg-hub.
          </DialogDescription>
        </DialogHeader>
        {generatedPeer ? (
          <div>
            <div>{JSON.stringify(generatedPeer, null, 2)}</div>
            <DialogFooter>
              <Button onClick={() => setOpen(false)}>Close</Button>
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
                  Add
                </Button>
              </DialogFooter>
            </form>
          </Form>
        )}
      </DialogContent>
    </Dialog>
  );
}
