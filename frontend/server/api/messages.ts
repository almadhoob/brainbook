import { sub } from 'date-fns'

const messages = [{
  id: 1,
  from: {
    name: 'Alice Smith',
    email: 'alice@uni.edu',
    avatar: {
      src: 'https://i.pravatar.cc/128?u=alice'
    }
  },
  subject: 'AI Ethics Paper Released',
  body: `Hi all,

I've just published my latest research paper on AI ethics. Would love your feedback and thoughts!

Best,
Alice`,
  date: new Date().toISOString()
}, {
  id: 2,
  unread: true,
  from: {
    name: 'Bob Johnson',
    email: 'bob@dev.com',
    avatar: {
      src: 'https://i.pravatar.cc/128?u=bobby'
    }
  },
  subject: 'Open Source Project Launch',
  body: `Hey team,

Check out my new open source project for web developers! Star it on GitHub and let me know your thoughts.

Cheers,
Bob`,
  date: sub(new Date(), { minutes: 7 }).toISOString()
}, {
  id: 3,
  unread: true,
  from: {
    name: 'Carol Lee',
    email: 'carol@data.org',
    avatar: {
      src: 'https://i.pravatar.cc/128?u=carol'
    }
  },
  subject: 'Data Visualization Notebook',
  body: `Hi everyone,

I've shared a Jupyter notebook for data visualization. Let me know if you find it helpful or have suggestions!

Best,
Carol`,
  date: sub(new Date(), { hours: 3 }).toISOString()
}, {
  id: 4,
  from: {
    name: 'David Nguyen',
    email: 'david@lab.net',
    avatar: {
      src: 'https://i.pravatar.cc/128?u=david'
    }
  },
  subject: 'ML Conference Slides',
  body: `Hi all,

Here are the slides from my recent ML conference talk. Happy to answer any questions or share code snippets.

Regards,
David`,
  date: sub(new Date(), { days: 1 }).toISOString()
}, {
  id: 5,
  from: {
    name: 'Eve Martins',
    email: 'eve@dev.com',
    avatar: {
      src: 'https://i.pravatar.cc/128?u=eve'
    }
  },
  subject: 'Vue & Nuxt Tips',
  body: `Hi devs,

Sharing some tips for building scalable Vue and Nuxt apps. Let me know your favorite tricks!

Best,
Eve`,
  date: sub(new Date(), { days: 1 }).toISOString()
}]

export default eventHandler(async () => {
  return messages
})
