// API route to get access token from secure httpOnly cookie
import type { NextApiRequest, NextApiResponse } from 'next';

export default function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== 'GET') {
    return res.status(405).json({ error: 'Method not allowed' });
  }

  try {
    const accessToken = req.cookies.kubechat_token;

    if (!accessToken) {
      return res.status(404).json({ error: 'No access token found' });
    }

    res.status(200).json({ accessToken });
  } catch (error) {
    console.error('Error getting access token:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}