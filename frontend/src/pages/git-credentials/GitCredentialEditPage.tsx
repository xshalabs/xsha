import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { ArrowLeft } from 'lucide-react';
import { usePageTitle } from '@/hooks/usePageTitle';
import { apiService } from '@/lib/api/index';
import { logError } from '@/lib/errors';
import { GitCredentialForm } from '@/components/GitCredentialForm';
import type { GitCredential } from '@/types/git-credentials';

const GitCredentialEditPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  
  const [credential, setCredential] = useState<GitCredential | null>(null);
  const [loading, setLoading] = useState(true);

  usePageTitle(credential ? `${t('gitCredentials.edit')} - ${credential.name}` : t('gitCredentials.edit'));

  // 加载凭据数据
  useEffect(() => {
    const loadCredential = async () => {
      if (!id) {
        logError(new Error('Credential ID is required'), 'Invalid credential ID');
        navigate('/git-credentials');
        return;
      }

      try {
        setLoading(true);
        const response = await apiService.gitCredentials.get(parseInt(id, 10));
        setCredential(response.credential);
      } catch (error) {
        logError(error as Error, 'Failed to load credential');
        alert(error instanceof Error ? error.message : t('gitCredentials.messages.loadFailed'));
        navigate('/git-credentials');
      } finally {
        setLoading(false);
      }
    };

    loadCredential();
  }, [id, navigate, t]);

  const handleSuccess = () => {
    navigate('/git-credentials');
  };

  const handleCancel = () => {
    navigate('/git-credentials');
  };

  if (loading) {
    return (
      <div className="container mx-auto p-6">
        <div className="max-w-2xl mx-auto">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
              <p className="text-muted-foreground">{t('common.loading')}</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!credential) {
    return null;
  }

  return (
    <div className="container mx-auto p-6">
      <div className="max-w-2xl mx-auto">
        <div className="mb-6">
          <Button 
            variant="outline" 
            onClick={() => navigate('/git-credentials')}
            className="mb-4"
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            {t('common.back')}
          </Button>
          <h1 className="text-2xl font-bold">{t('gitCredentials.edit')}</h1>
          <p className="text-muted-foreground mt-2">
            {t('gitCredentials.edit_description')} - {credential.name}
          </p>
        </div>

        <GitCredentialForm
          credential={credential}
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  );
};

export default GitCredentialEditPage; 