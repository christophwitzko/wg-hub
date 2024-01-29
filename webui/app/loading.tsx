import { Loader2 } from "lucide-react";
import { Center } from "@/components/center";

export default function Loading() {
  return (
    <Center>
      <Loader2 className="h-8 w-8 animate-spin" />
    </Center>
  );
}
