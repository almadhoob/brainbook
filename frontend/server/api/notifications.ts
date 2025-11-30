import { sub } from 'date-fns'

const notifications = [
  {
    id: 1,
    unread: true,
    sender: {
      name: 'Bob Johnson',
      email: 'bob@dev.com',
      avatar: { src: 'https://i.pravatar.cc/128?u=bobby' }
    },
    body: 'commented on your post',
    date: sub(new Date(), { minutes: 7 }).toISOString()
  },
  {
    id: 2,
    sender: {
      name: 'Eve Martins',
      email: 'eve@dev.com',
      avatar: { src: 'https://i.pravatar.cc/128?u=eve' }
    },
    body: 'invited you to join Web Developers group',
    date: sub(new Date(), { hours: 1 }).toISOString()
  },
  {
    id: 3,
    unread: true,
    sender: {
      name: 'Carol Lee',
      email: 'carol@data.org',
      avatar: { src: 'https://i.pravatar.cc/128?u=carol' }
    },
    body: 'shared a notebook with you',
    date: sub(new Date(), { hours: 3 }).toISOString()
  },
  {
    id: 4,
    sender: {
      name: 'David Nguyen',
      email: 'david@lab.net',
      avatar: { src: 'https://i.pravatar.cc/128?u=david' }
    },
    body: 'sent you a message',
    date: sub(new Date(), { hours: 7 }).toISOString()
  },
  {
    id: 5,
    sender: {
      name: 'Alice Smith',
      email: 'alice@uni.edu',
      avatar: { src: 'https://i.pravatar.cc/128?u=alice' }
    },
    body: 'added you to AI Researchers group',
    date: sub(new Date(), { days: 1 }).toISOString()
  },
  {
    id: 6,
    unread: true,
    sender: {
      name: 'Bob Johnson',
      email: 'bob@dev.com',
      avatar: { src: 'https://i.pravatar.cc/128?u=bobby' }
    },
    body: 'followed you',
    date: sub(new Date(), { days: 2 }).toISOString()
  },
  {
    id: 7,
    sender: {
      name: 'Eve Martins',
      email: 'eve@dev.com',
      avatar: { src: 'https://i.pravatar.cc/128?u=eve' }
    },
    body: 'commented on your post',
    date: sub(new Date(), { days: 5 }).toISOString()
  },
  {
    id: 8,
    sender: {
      name: 'Carol Lee',
      email: 'carol@data.org',
      avatar: { src: 'https://i.pravatar.cc/128?u=carol' }
    },
    body: 'joined Data Science Hub',
    date: sub(new Date(), { days: 6 }).toISOString()
  },
  {
    id: 9,
    sender: {
      name: 'David Nguyen',
      email: 'david@lab.net',
      avatar: { src: 'https://i.pravatar.cc/128?u=david' }
    },
    body: 'posted in AI Researchers group',
    date: sub(new Date(), { days: 7 }).toISOString()
  },
  {
    id: 10,
    sender: {
      name: 'Alice Smith',
      email: 'alice@uni.edu',
      avatar: { src: 'https://i.pravatar.cc/128?u=alice' }
    },
    body: 'shared an event: AI Ethics Webinar',
    date: sub(new Date(), { days: 9 }).toISOString()
  }
]

export default eventHandler(async () => {
  return notifications
})
