"use client";

import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

import {
  Card,
  CardHeader,
  CardFooter,
  CardContent,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { AlertCircle } from "lucide-react";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { useAuth } from "@/lib/auth";
import { Center } from "@/components/center";

const loginFormSchema = z.object({
  username: z.string().min(1).max(50),
  password: z.string().min(1).max(50),
});

export function Login() {
  const form = useForm<z.infer<typeof loginFormSchema>>({
    resolver: zodResolver(loginFormSchema),
    defaultValues: {
      username: "admin",
    },
  });
  const auth = useAuth();
  function onSubmit(values: z.infer<typeof loginFormSchema>) {
    auth.login(values.username, values.password);
  }

  return (
    <Form {...form}>
      <Center asChild={true}>
        <form onSubmit={form.handleSubmit(onSubmit)}>
          <Card className="w-1/4 min-w-80 m-4">
            <CardHeader className="text-2xl">Login</CardHeader>
            <CardContent className="grid w-full gap-4">
              <FormField
                control={form.control}
                name="username"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Username</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Password</FormLabel>
                    <FormControl>
                      <Input type="password" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              {auth.error && (
                <Alert variant="destructive">
                  <AlertCircle className="size-4" />
                  <AlertTitle>Error</AlertTitle>
                  <AlertDescription>{auth.error}</AlertDescription>
                </Alert>
              )}
            </CardContent>
            <CardFooter className="flex justify-end">
              <Button type="submit" disabled={auth.isLoading}>
                Log in
              </Button>
            </CardFooter>
          </Card>
        </form>
      </Center>
    </Form>
  );
}
