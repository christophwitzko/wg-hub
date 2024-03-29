"use client";

import { AlertCircle, Plus } from "lucide-react";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { Base64 } from "js-base64";

import { useAuth } from "@/lib/auth";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Form,
} from "@/components/ui/form";
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
import { Input } from "@/components/ui/input";
import { addPeer } from "@/lib/api";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { toast } from "sonner";

const addPeerFormSchema = z.object({
  publicKey: z.string().min(44).max(44).refine(Base64.isValid),
  allowedIP: z.string().ip({ version: "v4" }),
});

export function AddPeer() {
  const [open, setOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const form = useForm<z.infer<typeof addPeerFormSchema>>({
    resolver: zodResolver(addPeerFormSchema),
    defaultValues: {
      publicKey: "",
      allowedIP: "",
    },
  });
  const auth = useAuth();

  function onSubmit(values: z.infer<typeof addPeerFormSchema>) {
    setError("");
    setIsLoading(true);
    addPeer(auth.token, values)
      .then(() => {
        setOpen(false);
        toast("Peer added");
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
      setError("");
    }
  }, [form, open]);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline">
          Add Peer <Plus className="ml-2 size-4" />
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Add Peer</DialogTitle>
          <DialogDescription>Add a peer to wg-hub.</DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)}>
            <div className="grid gap-4 py-4">
              <FormField
                control={form.control}
                name="publicKey"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Public Key</FormLabel>
                    <FormControl>
                      <Input {...field} data-1p-ignore />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="allowedIP"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Allowed IP</FormLabel>
                    <FormControl>
                      <Input {...field} data-1p-ignore />
                    </FormControl>
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
      </DialogContent>
    </Dialog>
  );
}
