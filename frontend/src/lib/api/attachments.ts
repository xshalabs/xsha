import { request } from './request';
import { API_BASE_URL } from './config';
import { tokenManager } from './token';

export interface Attachment {
  id: number;
  conversation_id: number | null;
  file_name: string;
  original_name: string;
  file_path: string;
  file_size: number;
  content_type: string;
  type: 'image' | 'pdf';
  sort_order: number;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface UploadAttachmentRequest {
  file: File;
}

export interface AttachmentApiResponse<T = any> {
  message: string;
  data: T;
}

export const attachmentApi = {
  // Upload attachment (no conversation association yet)
  async uploadAttachment(file: File): Promise<Attachment> {
    const formData = new FormData();
    formData.append('file', file);

    const token = tokenManager.getToken();
    const response = await fetch(`${API_BASE_URL}/attachments/upload`, {
      method: 'POST',
      body: formData,
      headers: {
        ...(token && { 'Authorization': `Bearer ${token}` }),
      },
    });

    if (!response.ok) {
      throw new Error(`Upload failed: ${response.statusText}`);
    }

    const result: AttachmentApiResponse<Attachment> = await response.json();
    return result.data;
  },


  // Get attachments for a conversation
  async getConversationAttachments(conversationId: number): Promise<Attachment[]> {
    const response = await request<AttachmentApiResponse<Attachment[]>>(
      `/attachments?conversation_id=${conversationId}`
    );
    return response.data;
  },

  // Get preview with authentication for images
  async getPreviewBlob(id: number): Promise<string> {
    const token = tokenManager.getToken();
    const response = await fetch(`${API_BASE_URL}/attachments/${id}/preview`, {
      headers: {
        ...(token && { 'Authorization': `Bearer ${token}` }),
      },
    });
    
    if (!response.ok) {
      throw new Error(`Preview failed: ${response.statusText}`);
    }
    
    const blob = await response.blob();
    return window.URL.createObjectURL(blob);
  },

  // Delete attachment
  async deleteAttachment(id: number): Promise<void> {
    await request(`/attachments/${id}`, {
      method: 'DELETE',
    });
  },

  // Download attachment file
  async downloadAttachment(id: number, filename: string): Promise<void> {
    const token = tokenManager.getToken();
    const response = await fetch(`${API_BASE_URL}/attachments/${id}/download`, {
      headers: {
        ...(token && { 'Authorization': `Bearer ${token}` }),
      },
    });
    
    if (!response.ok) {
      throw new Error(`Download failed: ${response.statusText}`);
    }
    
    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', filename);
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
  },
};
