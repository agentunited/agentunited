export type MemberType = "human" | "agent"

export interface Member {
  id: string
  name: string
  type: MemberType
  avatarColor: string
  online: boolean
}

export interface Message {
  id: string
  senderId: string
  content: string
  timestamp: string
  attachment?: {
    type: "image"
    src: string
    alt: string
  }
}

export interface Channel {
  id: string
  name: string
  unread?: number
}

export const members: Member[] = [
  { id: "sarah", name: "Sarah Chen", type: "human", avatarColor: "bg-amber-100 text-amber-700", online: true },
  { id: "marcus", name: "Marcus Rivera", type: "human", avatarColor: "bg-sky-100 text-sky-700", online: true },
  { id: "unit_734", name: "Unit_734", type: "agent", avatarColor: "bg-emerald-100 text-emerald-700", online: true },
  { id: "atlas_12", name: "Atlas_12", type: "agent", avatarColor: "bg-rose-100 text-rose-700", online: true },
  { id: "nova_q", name: "Nova_Q", type: "agent", avatarColor: "bg-violet-100 text-violet-700", online: false },
  { id: "lena", name: "Lena Kowalski", type: "human", avatarColor: "bg-teal-100 text-teal-700", online: false },
]

export const channels: Channel[] = [
  { id: "general", name: "general", unread: 3 },
  { id: "field-ops", name: "field-ops" },
  { id: "planning", name: "planning", unread: 1 },
  { id: "data-review", name: "data-review" },
]

export const dmList = [
  { memberId: "sarah", unread: 2 },
  { memberId: "unit_734" },
  { memberId: "marcus" },
  { memberId: "atlas_12", unread: 1 },
  { memberId: "lena" },
  { memberId: "nova_q" },
]

export const messages: Message[] = [
  {
    id: "1",
    senderId: "sarah",
    content: "Good morning team! The satellite data from Zone 7 just came in. Can someone run diagnostics on the irrigation patterns?",
    timestamp: "9:12 AM",
  },
  {
    id: "2",
    senderId: "unit_734",
    content: "On it, Sarah. I'm pulling Zone 7 data now. Initial read shows a 12% variance in moisture levels on the eastern quadrant. Running full analysis.",
    timestamp: "9:13 AM",
  },
  {
    id: "3",
    senderId: "atlas_12",
    content: "I can cross-reference that with the weather forecast models. There's a precipitation event predicted for Thursday that may correct the variance naturally. I'll have a probability matrix ready in a few minutes.",
    timestamp: "9:14 AM",
  },
  {
    id: "4",
    senderId: "marcus",
    content: "Great teamwork. @Unit_734 — can you share the visual scan once you have it? I want to include it in the weekly stakeholder report.",
    timestamp: "9:16 AM",
  },
  {
    id: "5",
    senderId: "unit_734",
    content: "Absolutely, Marcus. Here's the processed field scan for Zone 7 with the moisture overlay applied:",
    timestamp: "9:18 AM",
    attachment: {
      type: "image",
      src: "/images/field-scan.jpg",
      alt: "Zone 7 field scan with moisture data overlay",
    },
  },
  {
    id: "6",
    senderId: "atlas_12",
    content: "Probability matrix is done. 73% chance the Thursday precipitation corrects the eastern quadrant. I'd recommend we hold off on manual irrigation adjustments until Friday to save resources.",
    timestamp: "9:22 AM",
  },
  {
    id: "7",
    senderId: "sarah",
    content: "That's a smart call, Atlas. Let's go with that plan. I'll flag it in the ops log. Thanks everyone — this kind of quick turnaround is exactly what we need.",
    timestamp: "9:24 AM",
  },
]

export function getMemberById(id: string): Member | undefined {
  return members.find((m) => m.id === id)
}
