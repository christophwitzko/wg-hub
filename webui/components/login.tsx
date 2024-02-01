"use client";

import {
  Card,
  CardHeader,
  CardFooter,
  CardContent,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { AlertCircle } from "lucide-react";
import { useState } from "react";
import { useAuth } from "@/lib/auth";
import { Center } from "@/components/center";

export function Login() {
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("");
  const auth = useAuth();

  const buttonDisabled = auth.isLoading || !username || !password;
  return (
    <Center asChild={true}>
      <form>
        <Card className="w-1/4 min-w-80 m-4">
          <CardHeader>Login</CardHeader>
          <CardContent className="grid w-full gap-4">
            <div className="flex flex-col space-y-1.5">
              <Label htmlFor="username">Username</Label>
              <Input
                id="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
              />
            </div>
            <div className="flex flex-col space-y-1.5">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
            {auth.error && (
              <Alert variant="destructive">
                <AlertCircle className="size-4" />
                <AlertTitle>Error</AlertTitle>
                <AlertDescription>{auth.error}</AlertDescription>
              </Alert>
            )}
          </CardContent>
          <CardFooter className="flex justify-end">
            <Button
              type="submit"
              disabled={buttonDisabled}
              onClick={() => auth.login(username, password)}
            >
              Log in
            </Button>
          </CardFooter>
        </Card>
      </form>
    </Center>
  );
}
