// API route to get refresh token from secure httpOnly cookie
import type { NextApiRequest, NextApiResponse } from 'next';

export default function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== 'GET') {
    return res.status(405).json({ error: 'Method not allowed' });
  }

  try {
    const refreshToken = req.cookies.kubechat_refresh_token;

    if (!refreshToken) {
      return res.status(404).json({ error: 'No refresh token found' });
    }

    res.status(200).json({ refreshToken });
  } catch (error) {
    console.error('Error getting refresh token:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}